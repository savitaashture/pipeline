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

package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewArtifactStorageFromMap(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		want    *ArtifactStorage
		wantErr bool
	}{
		{
			name: "all fields set",
			data: map[string]string{
				"oci-repository": "ghcr.io/org/project/artifacts",
				"insecure":       "true",
			},
			want: &ArtifactStorage{
				OCIRepository: "ghcr.io/org/project/artifacts",
				Insecure:      true,
			},
		},
		{
			name: "defaults when empty",
			data: map[string]string{},
			want: &ArtifactStorage{
				OCIRepository: "",
				Insecure:      false,
			},
		},
		{
			name: "only repository",
			data: map[string]string{
				"oci-repository": "registry:5000/artifacts",
			},
			want: &ArtifactStorage{
				OCIRepository: "registry:5000/artifacts",
				Insecure:      false,
			},
		},
		{
			name: "invalid insecure value",
			data: map[string]string{
				"insecure": "not-a-bool",
			},
			wantErr: true,
		},
		{
			// Regression test: an ConfigMap update that leaves a key present but
			// empty (e.g. a naive restore that writes "" instead of deleting the
			// key) must not fail config parsing -- a parse error here is fatal
			// and crashes the whole controller on config reload.
			name: "empty insecure value is treated as unset",
			data: map[string]string{
				"oci-repository": "registry:5000/artifacts",
				"insecure":       "",
			},
			want: &ArtifactStorage{
				OCIRepository: "registry:5000/artifacts",
				Insecure:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArtifactStorageFromMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewArtifactStorageFromMap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("mismatch (-want +got):\n%s", d)
			}
		})
	}
}

func TestNewArtifactStorageFromConfigMap(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "config-artifact-storage",
			Namespace: "tekton-pipelines",
		},
		Data: map[string]string{
			"oci-repository": "registry:5000/tekton-artifacts",
			"insecure":       "true",
		},
	}

	got, err := NewArtifactStorageFromConfigMap(cm)
	if err != nil {
		t.Fatalf("NewArtifactStorageFromConfigMap() error = %v", err)
	}
	want := &ArtifactStorage{
		OCIRepository: "registry:5000/tekton-artifacts",
		Insecure:      true,
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("mismatch (-want +got):\n%s", d)
	}
}
