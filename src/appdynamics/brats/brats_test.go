package brats_test

import (
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO The following test is pending because they currently fail.
// You need to make them pass by implementing the CopyBrats function in brats/brats_suite_test.go
var _ = PDescribe("appdynamics buildpack", func() {
	bratshelper.UnbuiltBuildpack("appdynamics", CopyBrats)
	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)
	bratshelper.StagingWithBuildpackThatSetsEOL("appdynamics", CopyBrats)
	bratshelper.StagingWithADepThatIsNotTheLatest("appdynamics", CopyBrats)
	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(`appdynamics\-[\d\.]+\-linux\-x64\-[\da-f]+\.tgz`, CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript("appdynamics", CopyBrats)
	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)
	bratshelper.ForAllSupportedVersions("appdynamics", CopyBrats, func(version string, app *cutlass.App) {
		PushApp(app)

		By("does a thing", func() {
			Expect(app).ToNot(BeNil())
		})
	})
})
