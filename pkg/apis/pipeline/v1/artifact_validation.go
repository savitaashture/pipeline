/*
Copyright 2026 The Tekton Authors

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
	"regexp"
	"strings"

	"github.com/tektoncd/pipeline/pkg/apis/config"
	"knative.dev/pkg/apis"
)

// stepArtifactValuePattern matches $(steps.<step>.artifacts.<name>)
var stepArtifactValuePattern = regexp.MustCompile(`^\$\(steps\.([^.]+)\.artifacts\.([^.)]+)\)$`)

// ValidateArtifactDeclarations validates artifact declarations on a TaskSpec.
func ValidateArtifactDeclarations(ctx context.Context, decls *ArtifactDeclarations) *apis.FieldError {
	if decls == nil {
		return nil
	}

	var errs *apis.FieldError

	cfg := config.FromContextOrDefaults(ctx)
	if cfg == nil || cfg.FeatureFlags == nil || !cfg.FeatureFlags.EnableArtifacts {
		return errs.Also(apis.ErrGeneric(fmt.Sprintf("feature flag %s should be set to true to use artifacts feature", config.EnableArtifacts), ""))
	}

	// Validate inputs
	inputNames := map[string]bool{}
	for i, input := range decls.Inputs {
		if input.Name == "" {
			errs = errs.Also(apis.ErrMissingField(fmt.Sprintf("inputs[%d].name", i)))
		} else if inputNames[input.Name] {
			errs = errs.Also(&apis.FieldError{
				Message: fmt.Sprintf("duplicate artifact input name %q", input.Name),
				Paths:   []string{fmt.Sprintf("inputs[%d].name", i)},
			})
		}
		inputNames[input.Name] = true
		if input.Type != "" && input.Type != ArtifactTypeContent && input.Type != ArtifactTypeReference {
			errs = errs.Also(&apis.FieldError{
				Message: fmt.Sprintf("invalid artifact type %q, must be %q or %q", input.Type, ArtifactTypeContent, ArtifactTypeReference),
				Paths:   []string{fmt.Sprintf("inputs[%d].type", i)},
			})
		}
		if input.Subject {
			errs = errs.Also(&apis.FieldError{
				Message: "subject is not allowed on artifact inputs, only on outputs",
				Paths:   []string{fmt.Sprintf("inputs[%d].subject", i)},
			})
		}
		if input.Value != "" {
			errs = errs.Also(&apis.FieldError{
				Message: "value is not allowed on artifact inputs, only on outputs",
				Paths:   []string{fmt.Sprintf("inputs[%d].value", i)},
			})
		}
	}

	// Validate outputs
	outputNames := map[string]bool{}
	for i, output := range decls.Outputs {
		if output.Name == "" {
			errs = errs.Also(apis.ErrMissingField(fmt.Sprintf("outputs[%d].name", i)))
		} else if outputNames[output.Name] {
			errs = errs.Also(&apis.FieldError{
				Message: fmt.Sprintf("duplicate artifact output name %q", output.Name),
				Paths:   []string{fmt.Sprintf("outputs[%d].name", i)},
			})
		}
		outputNames[output.Name] = true
		if output.Type != "" && output.Type != ArtifactTypeContent && output.Type != ArtifactTypeReference {
			errs = errs.Also(&apis.FieldError{
				Message: fmt.Sprintf("invalid artifact type %q, must be %q or %q", output.Type, ArtifactTypeContent, ArtifactTypeReference),
				Paths:   []string{fmt.Sprintf("outputs[%d].type", i)},
			})
		}
		if output.Value != "" && !stepArtifactValuePattern.MatchString(output.Value) {
			errs = errs.Also(&apis.FieldError{
				Message: fmt.Sprintf("invalid value %q, must match $(steps.<step>.artifacts.<name>)", output.Value),
				Paths:   []string{fmt.Sprintf("outputs[%d].value", i)},
			})
		}
	}

	return errs
}

// ValidatePipelineTaskArtifactBindings validates artifact bindings on a PipelineTask.
func ValidatePipelineTaskArtifactBindings(bindings *PipelineTaskArtifacts) *apis.FieldError {
	if bindings == nil {
		return nil
	}

	var errs *apis.FieldError
	for i, binding := range bindings.Inputs {
		if binding.Name == "" {
			errs = errs.Also(apis.ErrMissingField(fmt.Sprintf("inputs[%d].name", i)))
		}
		if binding.From == "" {
			errs = errs.Also(apis.ErrMissingField(fmt.Sprintf("inputs[%d].from", i)))
		} else {
			// Validate format: tasks.<taskName>.outputs.<artifactName>
			parts := strings.Split(binding.From, ".")
			if len(parts) != 4 || parts[0] != "tasks" || parts[2] != "outputs" {
				errs = errs.Also(&apis.FieldError{
					Message: fmt.Sprintf("invalid from format %q, expected 'tasks.<taskName>.outputs.<artifactName>'", binding.From),
					Paths:   []string{fmt.Sprintf("inputs[%d].from", i)},
				})
			}
		}
	}

	return errs
}
