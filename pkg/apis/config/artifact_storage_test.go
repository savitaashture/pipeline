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
				"enabled":                "true",
				"backend":                "oci",
				"inline-threshold":       "512",
				"oci-repository":         "ghcr.io/org/project/artifacts",
				"insecure":               "true",
				"oci.credentialsSecret":  "my-secret",
				"oci.attachReferrers":    "false",
				"oci.tagPattern":         "{{namespace}}-{{taskrun}}",
				"oci.groupByPipelineRun": "true",
			},
			want: &ArtifactStorage{
				Enabled:               true,
				Backend:               "oci",
				InlineThreshold:       512,
				OCIRepository:         "ghcr.io/org/project/artifacts",
				Insecure:              true,
				OCICredentialsSecret:  "my-secret",
				OCIAttachReferrers:    false,
				OCITagPattern:         "{{namespace}}-{{taskrun}}",
				OCIGroupByPipelineRun: true,
			},
		},
		{
			name: "defaults when empty",
			data: map[string]string{},
			want: &ArtifactStorage{
				Enabled:            false,
				Backend:            DefaultBackend,
				InlineThreshold:    DefaultInlineThreshold,
				OCIRepository:      "",
				Insecure:           false,
				OCIAttachReferrers: true,
				OCITagPattern:      DefaultTagPattern,
			},
		},
		{
			name: "only repository and enabled",
			data: map[string]string{
				"enabled":        "true",
				"oci-repository": "registry:5000/artifacts",
			},
			want: &ArtifactStorage{
				Enabled:            true,
				Backend:            DefaultBackend,
				InlineThreshold:    DefaultInlineThreshold,
				OCIRepository:      "registry:5000/artifacts",
				Insecure:           false,
				OCIAttachReferrers: true,
				OCITagPattern:      DefaultTagPattern,
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
			name: "invalid enabled value",
			data: map[string]string{
				"enabled": "not-a-bool",
			},
			wantErr: true,
		},
		{
			name: "invalid inline-threshold not a number",
			data: map[string]string{
				"inline-threshold": "abc",
			},
			wantErr: true,
		},
		{
			name: "invalid inline-threshold negative",
			data: map[string]string{
				"inline-threshold": "-1",
			},
			wantErr: true,
		},
		{
			name: "invalid inline-threshold exceeds ceiling",
			data: map[string]string{
				"inline-threshold": "4096",
			},
			wantErr: true,
		},
		{
			name: "inline-threshold at ceiling is valid",
			data: map[string]string{
				"inline-threshold": "2048",
			},
			want: &ArtifactStorage{
				Backend:            DefaultBackend,
				InlineThreshold:    2048,
				OCIAttachReferrers: true,
				OCITagPattern:      DefaultTagPattern,
			},
		},
		{
			name: "inline-threshold zero is valid",
			data: map[string]string{
				"inline-threshold": "0",
			},
			want: &ArtifactStorage{
				Backend:            DefaultBackend,
				InlineThreshold:    0,
				OCIAttachReferrers: true,
				OCITagPattern:      DefaultTagPattern,
			},
		},
		{
			name: "invalid oci.attachReferrers",
			data: map[string]string{
				"oci.attachReferrers": "not-a-bool",
			},
			wantErr: true,
		},
		{
			name: "invalid oci.groupByPipelineRun",
			data: map[string]string{
				"oci.groupByPipelineRun": "not-a-bool",
			},
			wantErr: true,
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
			"enabled":                "true",
			"backend":                "oci",
			"inline-threshold":       "1024",
			"oci-repository":         "registry:5000/tekton-artifacts",
			"insecure":               "true",
			"oci.credentialsSecret":  "my-registry-creds",
			"oci.attachReferrers":    "true",
			"oci.tagPattern":         "{{namespace}}.{{taskrun}}.{{artifact}}",
			"oci.groupByPipelineRun": "false",
		},
	}

	got, err := NewArtifactStorageFromConfigMap(cm)
	if err != nil {
		t.Fatalf("NewArtifactStorageFromConfigMap() error = %v", err)
	}
	want := &ArtifactStorage{
		Enabled:              true,
		Backend:              "oci",
		InlineThreshold:      1024,
		OCIRepository:        "registry:5000/tekton-artifacts",
		Insecure:             true,
		OCICredentialsSecret: "my-registry-creds",
		OCIAttachReferrers:   true,
		OCITagPattern:        "{{namespace}}.{{taskrun}}.{{artifact}}",
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("mismatch (-want +got):\n%s", d)
	}
}
