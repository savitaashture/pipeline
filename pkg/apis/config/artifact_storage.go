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

package config

import (
	"fmt"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

const (
	// ArtifactStorageConfigName is the name of the ConfigMap for artifact storage configuration.
	ArtifactStorageConfigName = "config-artifact-storage"

	enabledKey              = "enabled"
	backendKey              = "backend"
	inlineThresholdKey      = "inline-threshold"
	ociRepositoryKey        = "oci-repository"
	insecureKey             = "insecure"
	ociCredentialsSecretKey = "oci.credentialsSecret"
	ociAttachReferrersKey   = "oci.attachReferrers"
	ociTagPatternKey        = "oci.tagPattern"
	ociGroupByPipelineRun   = "oci.groupByPipelineRun"

	DefaultBackend         = "oci"
	DefaultInlineThreshold = 1024
	DefaultTagPattern      = "{{namespace}}.{{taskrun}}.{{artifact}}"

	// MaxInlineThreshold is the ceiling for inline-threshold (2KB),
	// leaving headroom in the per-container termination message budget.
	MaxInlineThreshold = 2048
)

// ArtifactStorage holds configuration for OCI-based artifact storage.
type ArtifactStorage struct {
	// Enabled controls whether content artifact storage is active.
	Enabled bool
	// Backend selects the storage backend type (e.g., "oci").
	Backend string
	// InlineThreshold is the size in bytes below which artifacts are stored
	// inline in TaskRun status rather than uploaded to the backend.
	InlineThreshold int
	// OCIRepository is the base OCI repository for artifact storage
	// e.g., "ghcr.io/org/project/artifacts" or "registry:5000/artifacts"
	OCIRepository string
	// Insecure allows plain HTTP for the OCI registry
	Insecure bool
	// OCICredentialsSecret is the name of a kubernetes.io/dockerconfigjson Secret
	// for OCI registry authentication. If empty, ServiceAccount imagePullSecrets are used.
	OCICredentialsSecret string
	// OCIAttachReferrers controls whether output artifacts are attached as
	// OCI referrers to subject artifacts after a PipelineRun completes.
	OCIAttachReferrers bool
	// OCITagPattern is the tag pattern for artifact manifests.
	// Variables: {{namespace}}, {{pipelinerun}}, {{taskrun}}, {{artifact}}
	OCITagPattern string
	// OCIGroupByPipelineRun creates a root OCI Index per PipelineRun and
	// attaches all artifacts as referrers to it.
	OCIGroupByPipelineRun bool
}

// NewArtifactStorageFromMap creates an ArtifactStorage from a map of string values.
func NewArtifactStorageFromMap(cfgMap map[string]string) (*ArtifactStorage, error) {
	as := &ArtifactStorage{
		Backend:            DefaultBackend,
		InlineThreshold:    DefaultInlineThreshold,
		OCIAttachReferrers: true,
		OCITagPattern:      DefaultTagPattern,
	}

	if v, ok := cfgMap[enabledKey]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing artifact storage config %q: %w for key %s", v, err, enabledKey)
		}
		as.Enabled = b
	}

	if v, ok := cfgMap[backendKey]; ok && v != "" {
		as.Backend = v
	}

	if v, ok := cfgMap[inlineThresholdKey]; ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing artifact storage config %q: %w for key %s", v, err, inlineThresholdKey)
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid artifact storage config: %s must be non-negative, got %d", inlineThresholdKey, n)
		}
		if n > MaxInlineThreshold {
			return nil, fmt.Errorf("invalid artifact storage config: %s must not exceed %d, got %d", inlineThresholdKey, MaxInlineThreshold, n)
		}
		as.InlineThreshold = n
	}

	if v, ok := cfgMap[ociRepositoryKey]; ok {
		as.OCIRepository = v
	}

	if v, ok := cfgMap[insecureKey]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing artifact storage config %q: %w for key %s", v, err, insecureKey)
		}
		as.Insecure = b
	}

	if v, ok := cfgMap[ociCredentialsSecretKey]; ok {
		as.OCICredentialsSecret = v
	}

	if v, ok := cfgMap[ociAttachReferrersKey]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing artifact storage config %q: %w for key %s", v, err, ociAttachReferrersKey)
		}
		as.OCIAttachReferrers = b
	}

	if v, ok := cfgMap[ociTagPatternKey]; ok && v != "" {
		as.OCITagPattern = v
	}

	if v, ok := cfgMap[ociGroupByPipelineRun]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing artifact storage config %q: %w for key %s", v, err, ociGroupByPipelineRun)
		}
		as.OCIGroupByPipelineRun = b
	}

	return as, nil
}

// NewArtifactStorageFromConfigMap creates an ArtifactStorage from a ConfigMap.
func NewArtifactStorageFromConfigMap(config *corev1.ConfigMap) (*ArtifactStorage, error) {
	return NewArtifactStorageFromMap(config.Data)
}

// GetArtifactStorageConfigName returns the name of the ConfigMap for artifact storage.
func GetArtifactStorageConfigName() string {
	if e := os.Getenv("CONFIG_ARTIFACT_STORAGE_NAME"); e != "" {
		return e
	}
	return ArtifactStorageConfigName
}
