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

package v1

import (
	"testing"
)

func TestValidateArtifactDeclarations_Valid(t *testing.T) {
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source"},
			{Name: "config", MediaType: "application/json"},
		},
		Outputs: []ArtifactDeclaration{
			{Name: "image", MediaType: "application/vnd.oci.image.manifest.v1+json", Subject: true},
			{Name: "sbom", MediaType: "application/vnd.cyclonedx+json"},
		},
	}

	if errs := ValidateArtifactDeclarations(decls); errs != nil {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_DuplicateInputNames(t *testing.T) {
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source"},
			{Name: "source"},
		},
	}

	errs := ValidateArtifactDeclarations(decls)
	if errs == nil {
		t.Error("expected error for duplicate input names")
	}
}

func TestValidateArtifactDeclarations_DuplicateOutputNames(t *testing.T) {
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image"},
			{Name: "image"},
		},
	}

	errs := ValidateArtifactDeclarations(decls)
	if errs == nil {
		t.Error("expected error for duplicate output names")
	}
}

func TestValidateArtifactDeclarations_EmptyName(t *testing.T) {
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: ""},
		},
	}

	errs := ValidateArtifactDeclarations(decls)
	if errs == nil {
		t.Error("expected error for empty artifact name")
	}
}

func TestValidateArtifactDeclarations_MultipleSubjectsAllowed(t *testing.T) {
	// Per TEP-0192, multiple artifacts can be marked as subject.
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image1", Subject: true},
			{Name: "image2", Subject: true},
		},
	}

	errs := ValidateArtifactDeclarations(decls)
	if errs != nil {
		t.Errorf("expected no error for multiple subjects, got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_InvalidType(t *testing.T) {
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image", Type: "bogus"},
		},
	}

	errs := ValidateArtifactDeclarations(decls)
	if errs == nil {
		t.Error("expected error for invalid artifact type")
	}
}

func TestValidateArtifactDeclarations_Nil(t *testing.T) {
	if errs := ValidateArtifactDeclarations(nil); errs != nil {
		t.Errorf("expected no errors for nil, got: %v", errs)
	}
}

func TestValidatePipelineTaskArtifactBindings_Valid(t *testing.T) {
	bindings := &PipelineTaskArtifacts{
		Inputs: []PipelineTaskArtifactBinding{
			{Name: "source", From: "tasks.build.outputs.image"},
		},
	}

	if errs := ValidatePipelineTaskArtifactBindings(bindings); errs != nil {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidatePipelineTaskArtifactBindings_InvalidFromFormat(t *testing.T) {
	bindings := &PipelineTaskArtifacts{
		Inputs: []PipelineTaskArtifactBinding{
			{Name: "source", From: "invalid-format"},
		},
	}

	errs := ValidatePipelineTaskArtifactBindings(bindings)
	if errs == nil {
		t.Error("expected error for invalid from format")
	}
}

func TestValidatePipelineTaskArtifactBindings_EmptyName(t *testing.T) {
	bindings := &PipelineTaskArtifacts{
		Inputs: []PipelineTaskArtifactBinding{
			{Name: "", From: "tasks.build.outputs.image"},
		},
	}

	errs := ValidatePipelineTaskArtifactBindings(bindings)
	if errs == nil {
		t.Error("expected error for empty binding name")
	}
}

func TestValidatePipelineTaskArtifactBindings_EmptyFrom(t *testing.T) {
	bindings := &PipelineTaskArtifacts{
		Inputs: []PipelineTaskArtifactBinding{
			{Name: "source", From: ""},
		},
	}

	errs := ValidatePipelineTaskArtifactBindings(bindings)
	if errs == nil {
		t.Error("expected error for empty from")
	}
}
