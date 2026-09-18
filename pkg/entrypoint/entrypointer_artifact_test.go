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
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func transportOpts(srv *httptest.Server) []remote.Option {
	return []remote.Option{remote.WithTransport(srv.Client().Transport)}
}

func TestEntrypointer_ArtifactOutputUpload(t *testing.T) {
	// Start an in-memory OCI registry
	reg := registry.New()
	srv := httptest.NewServer(reg)
	defer srv.Close()
	registryHost := strings.TrimPrefix(srv.URL, "http://")

	tmpDir := t.TempDir()
	stepMetadataDir := filepath.Join(tmpDir, "step-metadata")
	if err := os.MkdirAll(filepath.Join(stepMetadataDir, "artifacts"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create output artifact content (simulating what the step produces)
	outputDir := filepath.Join(tmpDir, "output-artifacts")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "sbom.json"), []byte(`{"components":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	e := Entrypointer{
		Command:         []string{"echo", "hello"},
		PostFile:        filepath.Join(tmpDir, "postfile"),
		Waiter:          &fakeWaiter{},
		Runner:          &fakeRunner{},
		PostWriter:      &fakePostWriter{},
		TerminationPath: filepath.Join(tmpDir, "termination"),
		StepMetadataDir: stepMetadataDir,
		ArtifactOutputs: []ArtifactOutput{
			{
				Name:       "sbom",
				Path:       outputDir,
				Repository: fmt.Sprintf("%s/artifacts/sbom", registryHost),
				MediaType:  "application/vnd.cyclonedx+json",
			},
		},
		ArtifactInsecure:     true,
		ArtifactRemoteOpts:   transportOpts(srv),
	}

	if err := e.Go(); err != nil {
		t.Fatalf("Entrypointer.Go() error = %v", err)
	}

	// Verify artifact results were written to termination message
	termContent, err := os.ReadFile(filepath.Join(tmpDir, "termination"))
	if err != nil {
		t.Fatalf("failed to read termination file: %v", err)
	}
	termStr := string(termContent)
	if !strings.Contains(termStr, "artifact-sbom") {
		t.Errorf("termination message should contain artifact-sbom, got: %s", termStr)
	}
	if !strings.Contains(termStr, "artifacts/sbom@sha256:") {
		t.Errorf("termination message should contain artifact URI, got: %s", termStr)
	}
}
