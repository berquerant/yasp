package internal

type Config struct {
	Debug         bool
	WorkDir       string
	IDTemplate    string
	K8SIDTemplate bool
	Cmds          []string
	Shell         string
	FailFast      bool
	FailFastCmd   bool
	Output        OutputMode
	Success       bool
	Kubectl       string
	Helm          string
	KustomizeRoot string
	HelmRoot      string
	Bulk          bool
}

type OutputMode int

const (
	OutputModeText OutputMode = iota
	OutputModeVerbose
	OutputModeYaml
)
