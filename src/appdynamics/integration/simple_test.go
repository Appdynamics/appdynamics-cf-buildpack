package integration_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple Integration Test", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	for _, version := range []string{"0.22.0", "0.24.2"} {
		Context("cystal version "+version, func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_test"))
				app.Path = CopyFixture(app.Path)

				config := make(map[string]interface{})
				Expect(libbuildpack.NewYAML().Load(filepath.Join(app.Path, "shard.yml"), config)).To(Succeed())
				config["appdynamics"] = version
				Expect(libbuildpack.NewYAML().Write(filepath.Join(app.Path, "shard.yml"), config)).To(Succeed())

				PushAppAndConfirm(app)
			})
			AfterEach(func() { os.RemoveAll(app.Path) })

			It("installs shards and runs", func() {
				Expect(log(app)).To(ContainSubstring("Installing appdynamics " + version))
				Expect(log(app)).To(ContainSubstring("Installing Dependencies"))

				Expect(app.GetBody("/")).To(ContainSubstring("IsBlank: true"))
			})
		})
	}
})
