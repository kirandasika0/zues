package config

// CurrentJobs hold all the in memory jobs
var CurrentJobs = map[string]Config{}

// Config defines a yaml configuration given by the developer
type Config struct {
	APIVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Type       string `yaml:"type,omitempty" json:"type,omitempty"`
	Spec       Spec   `yaml:"spec,omitempty" json:"spec,omitempty"`
}

//Spec is the global config spec for zues
type Spec struct {
	Name             string       `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace        string       `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	MaxBuildErrors   int32        `yaml:"maxBuildErrors,omitempty" json:"maxBuildErrors,omitempty"`
	MaxRetries       int32        `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty"`
	GatherStatistics bool         `yaml:"gatherStatistics,omitempty" json:"gatherStatistics,omitempty"`
	Notify           bool         `yaml:"notify,omitempty" json:"notify,omitempty"`
	StartupProbe     StartupProbe `yaml:"startupProbe,omitempty" json:"startupProbe,omitempty"`
	Rollback         Rollback     `yaml:"rollback,omitempty" json:"rollback,omitempty"`
}

// StartupProbe is a probe for startup
type StartupProbe struct {
	Spec StartupProbeSpec `yaml:"spec,omitempty" json:"spec,omitempty"`
}

// StartupProbeSpec represents specification described for zues to use while trying to set up Pods
type StartupProbeSpec struct {
	ServerType          string  `yaml:"serverType,omitempty" json:"serverType,omitempty"`
	Port                int16   `yaml:"port,omitempty" json:"port,omitempty"`
	Endpoint            string  `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	InitialStartupDelay int16   `yaml:"initialStartupDelay,omitempty" json:"initialStartupDelay,omitempty"`
	RetryThreshold      int16   `yaml:"retryThreshold,omitempty" json:"retryThreshold,omitempty"`
	ValidResponseCodes  []int16 `yaml:"validResponseCode,omitempty" json:"validResponseCode,omitempty"`
}

// Rollback is a struct
type Rollback struct {
	Spec RollbackSpec `yaml:"spec,omitempty" json:"spec,omitempty"`
}

// RollbackSpec is a struct
type RollbackSpec struct {
	RollbackImage string `yaml:"rollbackImage,omitempty" json:"rollbackImage,omitempty"`
	AdminNotify   bool   `yaml:"adminNotify,omitempty" json:"adminNotify,omitempty"`
}
