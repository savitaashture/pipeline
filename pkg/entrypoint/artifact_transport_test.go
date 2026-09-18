/*
Copyright 2024 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package entrypoint

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func TestUploadAndDownloadArtifact(t *testing.T) {
	// Start an in-memory OCI registry
	reg := registry.New()
	srv := httptest.NewServer(reg)
	defer srv.Close()

	registryHost := strings.TrimPrefix(srv.URL, "http://")

	// Create temp dir with test content to upload
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "result.json"), []byte(`{"passed": true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "nested.txt"), []byte("nested content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Upload
	output := ArtifactOutput{
		Name:       "test-results",
		Path:       srcDir,
		Repository: fmt.Sprintf("%s/artifacts/test-results", registryHost),
		MediaType:  "application/vnd.tekton.artifact.test.v1+tar.gz",
	}

	ctx := context.Background()
	av, err := UploadArtifact(ctx, output, true, remote.WithTransport(srv.Client().Transport))
	if err != nil {
		t.Fatalf("UploadArtifact() error = %v", err)
	}

	if av.Uri == "" {
		t.Error("expected non-empty URI")
	}
	if len(av.Digest) == 0 {
		t.Error("expected at least one digest entry")
	}
	if _, ok := av.Digest["sha256"]; !ok {
		t.Error("expected sha256 digest")
	}

	// Download to a new dir
	dstDir := t.TempDir()
	input := ArtifactInput{
		Name: "test-results",
		URI:  av.Uri,
		Path: dstDir,
	}

	if err := DownloadArtifact(ctx, input, true, remote.WithTransport(srv.Client().Transport)); err != nil {
		t.Fatalf("DownloadArtifact() error = %v", err)
	}

	// Verify files were extracted
	content, err := os.ReadFile(filepath.Join(dstDir, "result.json"))
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != `{"passed": true}` {
		t.Errorf("content mismatch: got %q", string(content))
	}

	nested, err := os.ReadFile(filepath.Join(dstDir, "sub", "nested.txt"))
	if err != nil {
		t.Fatalf("failed to read nested file: %v", err)
	}
	if string(nested) != "nested content" {
		t.Errorf("nested content mismatch: got %q", string(nested))
	}
}

func TestUploadArtifact_EmptyDir(t *testing.T) {
	reg := registry.New()
	srv := httptest.NewServer(reg)
	defer srv.Close()

	registryHost := strings.TrimPrefix(srv.URL, "http://")

	srcDir := t.TempDir()
	output := ArtifactOutput{
		Name:       "empty",
		Path:       srcDir,
		Repository: fmt.Sprintf("%s/artifacts/empty", registryHost),
		MediaType:  "application/vnd.tekton.artifact.v1+tar.gz",
	}

	ctx := context.Background()
	av, err := UploadArtifact(ctx, output, true, remote.WithTransport(srv.Client().Transport))
	if err != nil {
		t.Fatalf("UploadArtifact() error = %v", err)
	}
	if av.Uri == "" {
		t.Error("expected non-empty URI even for empty dir")
	}
}

func TestDownloadArtifact_InvalidURI(t *testing.T) {
	ctx := context.Background()
	input := ArtifactInput{
		Name: "bad",
		URI:  "not-a-valid-reference",
		Path: t.TempDir(),
	}
	err := DownloadArtifact(ctx, input, true)
	if err == nil {
		t.Error("expected error for invalid URI")
	}
}

func TestUploadArtifact_ReferenceType(t *testing.T) {
	srcDir := t.TempDir()
	uriFile := filepath.Join(srcDir, "image.uri")
	if err := os.WriteFile(uriFile, []byte("registry.example.com/myapp@sha256:"+strings.Repeat("a", 64)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output := ArtifactOutput{
		Name: "image",
		Type: "reference",
		Path: uriFile,
	}

	av, err := UploadArtifact(context.Background(), output, false)
	if err != nil {
		t.Fatalf("UploadArtifact() error = %v", err)
	}
	wantURI := "registry.example.com/myapp@sha256:" + strings.Repeat("a", 64)
	if av.Uri != wantURI {
		t.Errorf("Uri = %q, want %q", av.Uri, wantURI)
	}
	if av.Digest["sha256"] != strings.Repeat("a", 64) {
		t.Errorf("Digest[sha256] = %q, want %q", av.Digest["sha256"], strings.Repeat("a", 64))
	}
}

func TestUploadArtifact_ReferenceType_NoDigest(t *testing.T) {
	srcDir := t.TempDir()
	uriFile := filepath.Join(srcDir, "image.uri")
	if err := os.WriteFile(uriFile, []byte("registry.example.com/myapp:latest"), 0o644); err != nil {
		t.Fatal(err)
	}

	output := ArtifactOutput{
		Name: "image",
		Type: "reference",
		Path: uriFile,
	}

	if _, err := UploadArtifact(context.Background(), output, false); err == nil {
		t.Fatal("expected error for reference artifact missing a digest")
	}
}

func TestUploadArtifact_ContentType_NoRepository_DigestOnly(t *testing.T) {
	// TEP-0192: with content storage disabled (no repository configured),
	// content artifacts are still digested and recorded, just not uploaded.
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "data.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	output := ArtifactOutput{
		Name: "results",
		Type: "content",
		Path: srcDir,
		// Repository intentionally left empty.
	}

	av, err := UploadArtifact(context.Background(), output, false)
	if err != nil {
		t.Fatalf("UploadArtifact() error = %v", err)
	}
	if av.Uri != "" {
		t.Errorf("expected no Uri when storage is disabled, got %q", av.Uri)
	}
	if av.Digest["sha256"] == "" {
		t.Error("expected a sha256 digest even when storage is disabled")
	}
}
