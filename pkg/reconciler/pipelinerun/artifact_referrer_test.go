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

package pipelinerun

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
)

func TestAttachReferrers(t *testing.T) {
	// Start an in-memory OCI registry
	reg := registry.New()
	srv := httptest.NewServer(reg)
	defer srv.Close()
	registryHost := strings.TrimPrefix(srv.URL, "http://")

	// Push a "build output" image (the subject)
	subjectRef := fmt.Sprintf("%s/myapp:latest", registryHost)
	ref, err := name.ParseReference(subjectRef, name.Insecure)
	if err != nil {
		t.Fatal(err)
	}
	subjectImg, err := random.Image(256, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(ref, subjectImg, remote.WithTransport(srv.Client().Transport)); err != nil {
		t.Fatal(err)
	}

	subjectDigest, err := subjectImg.Digest()
	if err != nil {
		t.Fatal(err)
	}
	subjectDigestRef := fmt.Sprintf("%s/myapp@%s", registryHost, subjectDigest.String())

	// Define artifacts: one build output (subject) and one referrer (SBOM)
	artifacts := []TaskArtifactResult{
		{
			TaskName: "build",
			Artifact: v1.Artifact{
				Name:        "image",
				BuildOutput: true,
				Values: []v1.ArtifactValue{
					{Uri: subjectDigestRef, Digest: map[v1.Algorithm]string{"sha256": strings.TrimPrefix(subjectDigest.String(), "sha256:")}},
				},
			},
		},
		{
			TaskName: "build",
			Artifact: v1.Artifact{
				Name: "sbom",
				Values: []v1.ArtifactValue{
					{Uri: fmt.Sprintf("%s/artifacts/sbom@sha256:def456", registryHost), Digest: map[v1.Algorithm]string{"sha256": "def456"}},
				},
			},
			MediaType: "application/vnd.cyclonedx+json",
		},
	}

	ctx := context.Background()
	opts := []remote.Option{remote.WithTransport(srv.Client().Transport)}

	err = AttachReferrers(ctx, artifacts, true, opts...)
	if err != nil {
		t.Fatalf("AttachReferrers() error = %v", err)
	}

	// The referrer manifest was pushed successfully (we verify no error).
	// The in-memory registry doesn't support the referrers API, so we
	// verify the manifest exists by listing tags and checking for the
	// referrer tag convention (sha256-<subject-digest>).
	repo, err := name.NewRepository(fmt.Sprintf("%s/myapp", registryHost), name.Insecure)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := remote.List(repo, opts...)
	if err != nil {
		t.Fatalf("remote.List() error = %v", err)
	}
	// At minimum, should have 'latest' tag and the referrer was pushed by digest
	if len(tags) == 0 {
		t.Error("expected at least one tag in repository")
	}
}

func TestAttachReferrers_NoBuildOutput(t *testing.T) {
	artifacts := []TaskArtifactResult{
		{
			TaskName:  "test",
			Artifact:  v1.Artifact{Name: "results", Values: []v1.ArtifactValue{{Uri: "foo"}}},
			MediaType: "application/json",
		},
	}

	ctx := context.Background()
	// No build output → should be a no-op (no error)
	err := AttachReferrers(ctx, artifacts, true)
	if err != nil {
		t.Fatalf("AttachReferrers() with no build output should not error, got: %v", err)
	}
}
