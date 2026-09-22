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
	"context"
	"fmt"
	"testing"

	"github.com/tektoncd/pipeline/pkg/apis/config"
)

// artifactsEnabledCtx returns a context with enable-artifacts set to true.
func artifactsEnabledCtx() context.Context {
	return contextWithArtifactsEnabled(true)
}

// contextWithArtifactsEnabled returns a context with the given enable-artifacts value.
func contextWithArtifactsEnabled(enabled bool) context.Context {
	cfg := &config.Config{
		FeatureFlags: &config.FeatureFlags{
			EnableArtifacts: enabled,
		},
	}
	return config.ToContext(context.Background(), cfg)
}

func TestValidateArtifactDeclarations_FeatureFlagDisabled(t *testing.T) {
	ctx := contextWithArtifactsEnabled(false)
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image", Type: ArtifactTypeReference, Subject: true},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error when enable-artifacts is false")
	}
	want := fmt.Sprintf("feature flag %s should be set to true to use artifacts feature: ", config.EnableArtifacts)
	if errs.Error() != want {
		t.Errorf("expected error message %q, got %q", want, errs.Error())
	}
}

func TestValidateArtifactDeclarations_Valid(t *testing.T) {
	ctx := artifactsEnabledCtx()
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

	if errs := ValidateArtifactDeclarations(ctx, decls); errs != nil {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_DuplicateInputNames(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source"},
			{Name: "source"},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for duplicate input names")
	}
}

func TestValidateArtifactDeclarations_DuplicateOutputNames(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image"},
			{Name: "image"},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for duplicate output names")
	}
}

func TestValidateArtifactDeclarations_EmptyName(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: ""},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for empty artifact name")
	}
}

func TestValidateArtifactDeclarations_MultipleSubjectsAllowed(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image1", Subject: true},
			{Name: "image2", Subject: true},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs != nil {
		t.Errorf("expected no error for multiple subjects, got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_InvalidType(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image", Type: "bogus"},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for invalid artifact type")
	}
}

func TestValidateArtifactDeclarations_InvalidInputType(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source", Type: "bogus"},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for invalid input artifact type")
	}
}

func TestValidateArtifactDeclarations_ValidInputTypes(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source", Type: ArtifactTypeContent},
			{Name: "image-ref", Type: ArtifactTypeReference},
		},
	}

	if errs := ValidateArtifactDeclarations(ctx, decls); errs != nil {
		t.Errorf("expected no errors for valid input types, got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_SubjectOnInput(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source", Subject: true},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for subject on input")
	}
}

func TestValidateArtifactDeclarations_ValueOnInput(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Inputs: []ArtifactDeclaration{
			{Name: "source", Value: "$(steps.clone.artifacts.source)"},
		},
	}

	errs := ValidateArtifactDeclarations(ctx, decls)
	if errs == nil {
		t.Error("expected error for value on input")
	}
}

func TestValidateArtifactDeclarations_ValidOutputValue(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image", Value: "$(steps.build.artifacts.image)", Subject: true},
		},
	}

	if errs := ValidateArtifactDeclarations(ctx, decls); errs != nil {
		t.Errorf("expected no errors for valid output value, got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_InvalidOutputValueFormat(t *testing.T) {
	ctx := artifactsEnabledCtx()
	tests := []struct {
		name  string
		value string
	}{
		{"missing dollar-paren", "steps.build.artifacts.image"},
		{"wrong prefix", "$(tasks.build.artifacts.image)"},
		{"missing artifacts segment", "$(steps.build.image)"},
		{"extra segments", "$(steps.build.artifacts.image.uri)"},
		{"empty", "$(steps..artifacts.image)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decls := &ArtifactDeclarations{
				Outputs: []ArtifactDeclaration{
					{Name: "image", Value: tt.value},
				},
			}
			errs := ValidateArtifactDeclarations(ctx, decls)
			if errs == nil {
				t.Errorf("expected error for invalid value format %q", tt.value)
			}
		})
	}
}

func TestValidateArtifactDeclarations_TypeWithValueAllowed(t *testing.T) {
	ctx := artifactsEnabledCtx()
	decls := &ArtifactDeclarations{
		Outputs: []ArtifactDeclaration{
			{Name: "image", Type: ArtifactTypeContent, Value: "$(steps.build.artifacts.image)"},
		},
	}

	if errs := ValidateArtifactDeclarations(ctx, decls); errs != nil {
		t.Errorf("expected no error when type is set alongside value (type is defaulted), got: %v", errs)
	}
}

func TestValidateArtifactDeclarations_Nil(t *testing.T) {
	ctx := artifactsEnabledCtx()
	if errs := ValidateArtifactDeclarations(ctx, nil); errs != nil {
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
