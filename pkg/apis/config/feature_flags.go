/*
Copyright 2020 The Tekton Authors

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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	// StableAPIFields is the value used for "enable-api-fields" when only stable APIs should be usable.
	StableAPIFields = "stable"
	// AlphaAPIFields is the value used for "enable-api-fields" when alpha APIs should be usable as well.
	AlphaAPIFields = "alpha"
	// BetaAPIFields is the value used for "enable-api-fields" when beta APIs should be usable as well.
	BetaAPIFields = "beta"
	// FullEmbeddedStatus is the value used for "embedded-status" when the full statuses of TaskRuns and Runs should be
	// embedded in PipelineRunStatusFields, but ChildReferences should not be used.
	FullEmbeddedStatus = "full"
	// BothEmbeddedStatus is the value used for "embedded-status" when full embedded statuses of TaskRuns and Runs as
	// well as ChildReferences should be used in PipelineRunStatusFields.
	BothEmbeddedStatus = "both"
	// MinimalEmbeddedStatus is the value used for "embedded-status" when only ChildReferences should be used in
	// PipelineRunStatusFields.
	MinimalEmbeddedStatus = "minimal"
	// EnforceResourceVerificationMode is the value used for "resource-verification-mode" when verification is applied and fail the
	// TaskRun or PipelineRun when verification fails
	EnforceResourceVerificationMode = "enforce"
	// WarnResourceVerificationMode is the value used for "resource-verification-mode" when verification is applied but only log
	// the warning when verification fails
	WarnResourceVerificationMode = "warn"
	// SkipResourceVerificationMode is the value used for "resource-verification-mode" when verification is skipped
	SkipResourceVerificationMode = "skip"
	// ResultExtractionMethodTerminationMessage is the value used for "results-from" as a way to extract results from tasks using kubernetes termination message.
	ResultExtractionMethodTerminationMessage = "termination-message"
	// ResultExtractionMethodSidecarLogs is the value used for "results-from" as a way to extract results from tasks using sidecar logs.
	ResultExtractionMethodSidecarLogs = "sidecar-logs"
	// CustomTaskVersionAlpha is the value used for "custom-task-version" when the PipelineRun reconciler should create
	// v1alpha1.Runs.
	CustomTaskVersionAlpha = "v1alpha1"
	// CustomTaskVersionBeta is the value used for "custom-task-version" when the PipelineRun reconciler should create
	// v1beta1.CustomRuns.
	CustomTaskVersionBeta = "v1beta1"
	// DefaultDisableAffinityAssistant is the default value for "disable-affinity-assistant".
	DefaultDisableAffinityAssistant = false
	// DefaultDisableCredsInit is the default value for "disable-creds-init".
	DefaultDisableCredsInit = false
	// DefaultRunningInEnvWithInjectedSidecars is the default value for "running-in-environment-with-injected-sidecars".
	DefaultRunningInEnvWithInjectedSidecars = true
	// DefaultAwaitSidecarReadiness is the default value for "await-sidecar-readiness".
	DefaultAwaitSidecarReadiness = true
	// DefaultRequireGitSSHSecretKnownHosts is the default value for "require-git-ssh-secret-known-hosts".
	DefaultRequireGitSSHSecretKnownHosts = false
	// DefaultEnableTektonOciBundles is the default value for "enable-tekton-oci-bundles".
	DefaultEnableTektonOciBundles = false
	// DefaultEnableAPIFields is the default value for "enable-api-fields".
	DefaultEnableAPIFields = StableAPIFields
	// DefaultSendCloudEventsForRuns is the default value for "send-cloudevents-for-runs".
	DefaultSendCloudEventsForRuns = false
	// DefaultEmbeddedStatus is the default value for "embedded-status".
	DefaultEmbeddedStatus = MinimalEmbeddedStatus
	// EnforceNonfalsifiabilityWithSpire is the value used for  "enable-nonfalsifiability" when SPIRE is used to enable non-falsifiability.
	EnforceNonfalsifiabilityWithSpire = "spire"
	// EnforceNonfalsifiabilityNone is the value used for  "enable-nonfalsifiability" when non-falsifiability is not enabled.
	EnforceNonfalsifiabilityNone = ""
	// DefaultEnforceNonfalsifiability is the default value for "enforce-nonfalsifiability".
	DefaultEnforceNonfalsifiability = EnforceNonfalsifiabilityNone
	// DefaultResourceVerificationMode is the default value for "resource-verification-mode".
	DefaultResourceVerificationMode = SkipResourceVerificationMode
	// DefaultEnableProvenanceInStatus is the default value for "enable-provenance-status".
	DefaultEnableProvenanceInStatus = false
	// DefaultResultExtractionMethod is the default value for ResultExtractionMethod
	DefaultResultExtractionMethod = ResultExtractionMethodTerminationMessage
	// DefaultMaxResultSize is the default value in bytes for the size of a result
	DefaultMaxResultSize = 4096
	// DefaultCustomTaskVersion is the default value for "custom-task-version"
	DefaultCustomTaskVersion = CustomTaskVersionBeta

	disableAffinityAssistantKey         = "disable-affinity-assistant"
	disableCredsInitKey                 = "disable-creds-init"
	runningInEnvWithInjectedSidecarsKey = "running-in-environment-with-injected-sidecars"
	awaitSidecarReadinessKey            = "await-sidecar-readiness"
	requireGitSSHSecretKnownHostsKey    = "require-git-ssh-secret-known-hosts" // nolint: gosec
	enableTektonOCIBundles              = "enable-tekton-oci-bundles"
	enableAPIFields                     = "enable-api-fields"
	sendCloudEventsForRuns              = "send-cloudevents-for-runs"
	embeddedStatus                      = "embedded-status"
	enforceNonfalsifiability            = "enforce-nonfalsifiability"
	verificationMode                    = "resource-verification-mode"
	enableProvenanceInStatus            = "enable-provenance-in-status"
	resultExtractionMethod              = "results-from"
	maxResultSize                       = "max-result-size"
	customTaskVersion                   = "custom-task-version"
)

// FeatureFlags holds the features configurations
// +k8s:deepcopy-gen=true
type FeatureFlags struct {
	DisableAffinityAssistant         bool
	DisableCredsInit                 bool
	RunningInEnvWithInjectedSidecars bool
	RequireGitSSHSecretKnownHosts    bool
	EnableTektonOCIBundles           bool
	ScopeWhenExpressionsToTask       bool
	EnableAPIFields                  string
	SendCloudEventsForRuns           bool
	AwaitSidecarReadiness            bool
	EmbeddedStatus                   string
	EnforceNonfalsifiability         string
	ResourceVerificationMode         string
	EnableProvenanceInStatus         bool
	ResultExtractionMethod           string
	MaxResultSize                    int
	CustomTaskVersion                string
}

// GetFeatureFlagsConfigName returns the name of the configmap containing all
// feature flags.
func GetFeatureFlagsConfigName() string {
	if e := os.Getenv("CONFIG_FEATURE_FLAGS_NAME"); e != "" {
		return e
	}
	return "feature-flags"
}

func getEnforceNonfalsifiabilityFeature(cfgMap map[string]string) (string, error) {
	var mapValue struct{}
	var acceptedValues = map[string]struct{}{
		EnforceNonfalsifiabilityNone:      mapValue,
		EnforceNonfalsifiabilityWithSpire: mapValue,
	}
	var value = DefaultEnforceNonfalsifiability
	if cfg, ok := cfgMap[enforceNonfalsifiability]; ok {
		value = strings.ToLower(cfg)
	}
	if _, ok := acceptedValues[value]; !ok {
		return DefaultEnforceNonfalsifiability, fmt.Errorf("invalid value for feature flag %q: %q", enforceNonfalsifiability, value)
	}
	return value, nil
}

// NewFeatureFlagsFromMap returns a Config given a map corresponding to a ConfigMap
func NewFeatureFlagsFromMap(cfgMap map[string]string) (*FeatureFlags, error) {
	setFeature := func(key string, defaultValue bool, feature *bool) error {
		if cfg, ok := cfgMap[key]; ok {
			value, err := strconv.ParseBool(cfg)
			if err != nil {
				return fmt.Errorf("failed parsing feature flags config %q: %v", cfg, err)
			}
			*feature = value
			return nil
		}
		*feature = defaultValue
		return nil
	}

	tc := FeatureFlags{}
	if err := setFeature(disableAffinityAssistantKey, DefaultDisableAffinityAssistant, &tc.DisableAffinityAssistant); err != nil {
		return nil, err
	}
	if err := setFeature(disableCredsInitKey, DefaultDisableCredsInit, &tc.DisableCredsInit); err != nil {
		return nil, err
	}
	if err := setFeature(runningInEnvWithInjectedSidecarsKey, DefaultRunningInEnvWithInjectedSidecars, &tc.RunningInEnvWithInjectedSidecars); err != nil {
		return nil, err
	}
	if err := setFeature(awaitSidecarReadinessKey, DefaultAwaitSidecarReadiness, &tc.AwaitSidecarReadiness); err != nil {
		return nil, err
	}
	if err := setFeature(requireGitSSHSecretKnownHostsKey, DefaultRequireGitSSHSecretKnownHosts, &tc.RequireGitSSHSecretKnownHosts); err != nil {
		return nil, err
	}
	if err := setEnabledAPIFields(cfgMap, DefaultEnableAPIFields, &tc.EnableAPIFields); err != nil {
		return nil, err
	}
	if err := setFeature(sendCloudEventsForRuns, DefaultSendCloudEventsForRuns, &tc.SendCloudEventsForRuns); err != nil {
		return nil, err
	}
	if err := setEmbeddedStatus(cfgMap, DefaultEmbeddedStatus, &tc.EmbeddedStatus); err != nil {
		return nil, err
	}
	if err := setResourceVerificationMode(cfgMap, DefaultResourceVerificationMode, &tc.ResourceVerificationMode); err != nil {
		return nil, err
	}
	if err := setFeature(enableProvenanceInStatus, DefaultEnableProvenanceInStatus, &tc.EnableProvenanceInStatus); err != nil {
		return nil, err
	}
	if err := setResultExtractionMethod(cfgMap, DefaultResultExtractionMethod, &tc.ResultExtractionMethod); err != nil {
		return nil, err
	}
	if err := setMaxResultSize(cfgMap, DefaultMaxResultSize, &tc.MaxResultSize); err != nil {
		return nil, err
	}
	if err := setCustomTaskVersion(cfgMap, DefaultCustomTaskVersion, &tc.CustomTaskVersion); err != nil {
		return nil, err
	}

	// Given that they are alpha features, Tekton Bundles and Custom Tasks should be switched on if
	// enable-api-fields is "alpha". If enable-api-fields is not "alpha" then fall back to the value of
	// each feature's individual flag.
	//
	// Note: the user cannot enable "alpha" while disabling bundles or custom tasks - that would
	// defeat the purpose of having a single shared gate for all alpha features.
	if tc.EnableAPIFields == AlphaAPIFields {
		tc.EnableTektonOCIBundles = true
		// Only consider SPIRE if alpha is on.
		enforceNonfalsifiabilityValue, err := getEnforceNonfalsifiabilityFeature(cfgMap)
		if err != nil {
			return nil, err
		}
		tc.EnforceNonfalsifiability = enforceNonfalsifiabilityValue
	} else {
		if err := setFeature(enableTektonOCIBundles, DefaultEnableTektonOciBundles, &tc.EnableTektonOCIBundles); err != nil {
			return nil, err
		}
		// Do not enable any form of non-falsifiability enforcement in non-alpha mode.
		tc.EnforceNonfalsifiability = EnforceNonfalsifiabilityNone
		if enforceNonfalsifiabilityValue, err := getEnforceNonfalsifiabilityFeature(cfgMap); err != nil || enforceNonfalsifiabilityValue != DefaultEnforceNonfalsifiability {
			return nil, fmt.Errorf("%q can be set to non-default values (%q) only in alpha", enforceNonfalsifiability, enforceNonfalsifiabilityValue)
		}
	}
	return &tc, nil
}

// setEnabledAPIFields sets the "enable-api-fields" flag based on the content of a given map.
// If the feature gate is invalid or missing then an error is returned.
func setEnabledAPIFields(cfgMap map[string]string, defaultValue string, feature *string) error {
	value := defaultValue
	if cfg, ok := cfgMap[enableAPIFields]; ok {
		value = strings.ToLower(cfg)
	}
	switch value {
	case AlphaAPIFields, BetaAPIFields, StableAPIFields:
		*feature = value
	default:
		return fmt.Errorf("invalid value for feature flag %q: %q", enableAPIFields, value)
	}
	return nil
}

// setEmbeddedStatus sets the "embedded-status" flag based on the content of a given map.
// If the feature gate is invalid or missing then an error is returned.
func setEmbeddedStatus(cfgMap map[string]string, defaultValue string, feature *string) error {
	value := defaultValue
	if cfg, ok := cfgMap[embeddedStatus]; ok {
		value = strings.ToLower(cfg)
	}
	switch value {
	case FullEmbeddedStatus, BothEmbeddedStatus, MinimalEmbeddedStatus:
		*feature = value
	default:
		return fmt.Errorf("invalid value for feature flag %q: %q", embeddedStatus, value)
	}
	return nil
}

// setResultExtractionMethod sets the "results-from" flag based on the content of a given map.
// If the feature gate is invalid or missing then an error is returned.
func setResultExtractionMethod(cfgMap map[string]string, defaultValue string, feature *string) error {
	value := defaultValue
	if cfg, ok := cfgMap[resultExtractionMethod]; ok {
		value = strings.ToLower(cfg)
	}
	switch value {
	case ResultExtractionMethodTerminationMessage, ResultExtractionMethodSidecarLogs:
		*feature = value
	default:
		return fmt.Errorf("invalid value for feature flag %q: %q", resultExtractionMethod, value)
	}
	return nil
}

// setCustomTaskVersion sets the "custom-task-version" flag based on the content of a given map.
// If the feature gate is invalid or missing then an error is returned.
func setCustomTaskVersion(cfgMap map[string]string, defaultValue string, feature *string) error {
	value := defaultValue
	if cfg, ok := cfgMap[customTaskVersion]; ok {
		value = strings.ToLower(cfg)
	}
	switch value {
	case CustomTaskVersionAlpha, CustomTaskVersionBeta:
		*feature = value
	default:
		return fmt.Errorf("invalid value for feature flag %q: %q", customTaskVersion, value)
	}
	return nil
}

// setMaxResultSize sets the "max-result-size" flag based on the content of a given map.
// If the feature gate is invalid or missing then an error is returned.
func setMaxResultSize(cfgMap map[string]string, defaultValue int, feature *int) error {
	value := defaultValue
	if cfg, ok := cfgMap[maxResultSize]; ok {
		v, err := strconv.Atoi(cfg)
		if err != nil {
			return err
		}
		value = v
	}
	// if max limit is > 1.5 MB (CRD limit).
	if value >= 1572864 {
		return fmt.Errorf("invalid value for feature flag %q: %q. This is exceeding the CRD limit", resultExtractionMethod, value)
	}
	*feature = value
	return nil
}

// setResourceVerificationMode sets the "resource-verification-mode" flag based on the content of a given map.
// If the value is invalid or missing then an error is returned.
func setResourceVerificationMode(cfgMap map[string]string, defaultValue string, feature *string) error {
	value := defaultValue
	if cfg, ok := cfgMap[verificationMode]; ok {
		value = strings.ToLower(cfg)
	}
	switch value {
	case EnforceResourceVerificationMode, WarnResourceVerificationMode, SkipResourceVerificationMode:
		*feature = value
	default:
		return fmt.Errorf("invalid value for feature flag %q: %q", verificationMode, value)
	}
	return nil
}

// NewFeatureFlagsFromConfigMap returns a Config for the given configmap
func NewFeatureFlagsFromConfigMap(config *corev1.ConfigMap) (*FeatureFlags, error) {
	return NewFeatureFlagsFromMap(config.Data)
}

// EnableAlphaAPIFields enables alpha features in an existing context (for use in testing)
func EnableAlphaAPIFields(ctx context.Context) context.Context {
	return setEnableAPIFields(ctx, "alpha")
}

// EnableBetaAPIFields enables beta features in an existing context (for use in testing)
func EnableBetaAPIFields(ctx context.Context) context.Context {
	return setEnableAPIFields(ctx, "beta")
}

// CheckEnforceResourceVerificationMode returns true if the ResourceVerificationMode is EnforceResourceVerificationMode
// else returns false
func CheckEnforceResourceVerificationMode(ctx context.Context) bool {
	cfg := FromContextOrDefaults(ctx)
	return cfg.FeatureFlags.ResourceVerificationMode == EnforceResourceVerificationMode
}

// CheckWarnResourceVerificationMode returns true if the ResourceVerificationMode is WarnResourceVerificationMode
// else returns false
func CheckWarnResourceVerificationMode(ctx context.Context) bool {
	cfg := FromContextOrDefaults(ctx)
	return cfg.FeatureFlags.ResourceVerificationMode == WarnResourceVerificationMode
}

func setEnableAPIFields(ctx context.Context, want string) context.Context {
	featureFlags, _ := NewFeatureFlagsFromMap(map[string]string{
		"enable-api-fields": want,
	})
	cfg := &Config{
		Defaults: &Defaults{
			DefaultTimeoutMinutes: 60,
		},
		FeatureFlags: featureFlags,
	}
	return ToContext(ctx, cfg)
}
