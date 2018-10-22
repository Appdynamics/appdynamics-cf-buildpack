package finalize

import (
	"fmt"
	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
}

type Finalizer struct {
	Stager Stager
	Log    *libbuildpack.Logger
}

func (f *Finalizer) Run() error {
	f.Log.BeginStep("Configuring appdynamics")

	data, err := f.GenerateReleaseYaml()
	if err != nil {
		f.Log.Error("Error generating release YAML: %v", err)
		return err
	}
	libbuildpack.NewYAML().Write("/tmp/appdynamics-buildpack-release-step.yml", data)

	return nil
}

func (f *Finalizer) GenerateReleaseYaml() (map[string]map[string]string, error) {
	processTypes := map[string]string{
		"web": fmt.Sprintf("$DEPS_DIR/%s/app --port $PORT", f.Stager.DepsIdx()),
	}
	return map[string]map[string]string{
		"default_process_types": processTypes,
	}, nil
}
