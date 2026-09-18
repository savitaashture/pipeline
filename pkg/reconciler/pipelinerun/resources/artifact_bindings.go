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

package resources

import (
	"fmt"
	"path/filepath"
	"strings"

	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/tektoncd/pipeline/pkg/entrypoint"
)

// ArtifactInputsAnnotation is the annotation key for passing resolved artifact input URIs
// from the pipeline reconciler to the pod builder.
const ArtifactInputsAnnotation = "tekton.dev/artifact-inputs"

// ResolveArtifactBinding resolves a "from" reference (e.g., "tasks.build.outputs.image")
// to the URI of the artifact from completed TaskRun artifacts.
// The artifacts map is keyed by PipelineTask name.
func ResolveArtifactBinding(from string, artifacts map[string]*v1.Artifacts) (string, error) {
	// Expected format: tasks.<taskName>.outputs.<artifactName>
	parts := strings.Split(from, ".")
	if len(parts) != 4 || parts[0] != "tasks" || parts[2] != "outputs" {
		return "", fmt.Errorf("invalid artifact binding format %q, expected 'tasks.<taskName>.outputs.<artifactName>'", from)
	}

	taskName := parts[1]
	artifactName := parts[3]

	taskArtifacts, ok := artifacts[taskName]
	if !ok {
		return "", fmt.Errorf("task %q not found in completed artifacts", taskName)
	}

	for _, a := range taskArtifacts.Outputs {
		if a.Name == artifactName {
			if len(a.Values) == 0 {
				return "", fmt.Errorf("artifact %q from task %q has no values", artifactName, taskName)
			}
			// Return the first (most recent) value's URI
			return a.Values[0].Uri, nil
		}
	}

	return "", fmt.Errorf("artifact %q not found in task %q outputs", artifactName, taskName)
}

// ResolveArtifactInputsForTask resolves all artifact input bindings for a PipelineTask,
// returning ArtifactInput structs ready to be serialized as entrypoint args.
func ResolveArtifactInputsForTask(pt *v1.PipelineTask, taskArtifacts map[string]*v1.Artifacts) ([]entrypoint.ArtifactInput, error) {
	if pt.Artifacts == nil || len(pt.Artifacts.Inputs) == 0 {
		return nil, nil
	}

	var inputs []entrypoint.ArtifactInput
	for _, binding := range pt.Artifacts.Inputs {
		uri, err := ResolveArtifactBinding(binding.From, taskArtifacts)
		if err != nil {
			return nil, fmt.Errorf("resolving artifact input %q: %w", binding.Name, err)
		}
		inputs = append(inputs, entrypoint.ArtifactInput{
			Name: binding.Name,
			URI:  uri,
			Path: filepath.Join("/workspace/artifacts/inputs", binding.Name),
		})
	}

	return inputs, nil
}
