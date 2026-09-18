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

package resources_test

import (
	"testing"

	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	resources "github.com/tektoncd/pipeline/pkg/reconciler/taskrun/resources"
)

func TestApplyArtifactDeclarationPaths(t *testing.T) {
	spec := &v1.TaskSpec{
		Artifacts: &v1.ArtifactDeclarations{
			Inputs: []v1.ArtifactDeclaration{
				{Name: "source"},
			},
			Outputs: []v1.ArtifactDeclaration{
				{Name: "results"},
			},
		},
		Steps: []v1.Step{{
			Name:  "test",
			Image: "busybox",
			Script: `
cd $(inputs.source.path)
mkdir -p $(outputs.results.path)
echo done > $(outputs.results.path)/data.json
`,
		}},
	}

	got := resources.ApplyArtifactDeclarationPaths(spec)

	expectedScript := `
cd /workspace/artifacts/inputs/source
mkdir -p /workspace/artifacts/outputs/results
echo done > /workspace/artifacts/outputs/results/data.json
`
	if got.Steps[0].Script != expectedScript {
		t.Errorf("script mismatch:\ngot:  %q\nwant: %q", got.Steps[0].Script, expectedScript)
	}
}

func TestApplyArtifactDeclarationPaths_NoArtifacts(t *testing.T) {
	spec := &v1.TaskSpec{
		Steps: []v1.Step{{
			Name:   "test",
			Image:  "busybox",
			Script: "echo hello",
		}},
	}

	got := resources.ApplyArtifactDeclarationPaths(spec)
	if got.Steps[0].Script != "echo hello" {
		t.Errorf("script should not change without artifacts, got: %q", got.Steps[0].Script)
	}
}
