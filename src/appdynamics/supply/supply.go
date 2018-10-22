package supply

import (
	"github.com/cloudfoundry/libbuildpack"
	"os/exec"
)

type Stager interface {
	BuildDir() string
	CacheDir() string
	DepDir() string
	LinkDirectoryInDepDir(string, string) error
}

type Manifest interface {
	RootDir() string
	AllDependencyVersions(string) []string
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Command interface {
	Output(dir string, program string, args ...string) (string, error)
	Run(cmd *exec.Cmd) error
}

type Supplier struct {
	Manifest Manifest
	Stager   Stager
	Command  Command
	Log      *libbuildpack.Logger
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying appdynamics")
	return nil
}

