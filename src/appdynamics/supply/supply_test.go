package supply_test

import (
	"bytes"
	"appdynamics/supply"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supply.go --destination=mocks_test.go --package=supply_test

var _ = Describe("Supply", func() {
	var (
		err          error
		buildDir     string
		cacheDir     string
		depsDir      string
		depsIdx      string
		supplier     *supply.Supplier
		logger       *libbuildpack.Logger
		buffer       *bytes.Buffer
		mockCtrl     *gomock.Controller
		mockManifest *MockManifest
		mockCommand  *MockCommand
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "php-buildpack.build.")
		Expect(err).To(BeNil())
		cacheDir, err = ioutil.TempDir("", "php-buildpack.cache.")
		Expect(err).To(BeNil())
		depsDir, err = ioutil.TempDir("", "php-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "9"
		Expect(os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)).To(Succeed())

		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)

		args := []string{buildDir, cacheDir, depsDir, depsIdx}
		stager := libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})

		supplier = &supply.Supplier{
			Stager:   stager,
			Manifest: mockManifest,
			Log:      logger,
			Command:  mockCommand,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
		Expect(os.RemoveAll(buildDir)).To(Succeed())
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
		Expect(os.RemoveAll(depsDir)).To(Succeed())
	})

	Describe("Setup", func() {
		BeforeEach(func() {
			mockManifest.EXPECT().AllDependencyVersions("appdynamics").Return([]string{"0.21.1", "0.22.3", "0.23.4"})
		})
		Context("shard.yml exists in builddir", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "shard.yml"), []byte(`{"name":"appname"}`), 0644)).To(Succeed())
			})
			It("sets Shard.Name", func() {
				Expect(supplier.Setup()).To(Succeed())
				Expect(supplier.Shard.Name).To(Equal("appname"))
			})
		})
		Context("shard.yml does NOT exist in builddir", func() {
			It("Shard.Name is blank", func() {
				Expect(supplier.Setup()).To(Succeed())
				Expect(supplier.Shard.Name).To(Equal(""))
			})
		})

		Context("user sets CRYSTAL_VERSION", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(buildDir, "shard.yml"), []byte(`{"name":"appname","appdynamics":"0.22.x"}`), 0644)).To(Succeed())
			})
			It("interpolates the version", func() {
				Expect(supplier.Setup()).To(Succeed())
				Expect(supplier.Shard.CrystalVersion).To(Equal("0.22.3"))
			})
		})

		Context("user does NOT set CRYSTAL_VERSION", func() {
			BeforeEach(func() {
				supplier.Shard.CrystalVersion = ""
			})
			It("sets version to highest available", func() {
				Expect(supplier.Setup()).To(Succeed())
				Expect(supplier.Shard.CrystalVersion).To(Equal("0.23.4"))
			})
		})
	})

	PIt("UntarLibevent", func() {})

	Describe("InstallCrystal", func() {
		BeforeEach(func() {
			supplier.Shard.CrystalVersion = "0.23.4"
			mockManifest.EXPECT().InstallDependency(libbuildpack.Dependency{Name: "appdynamics", Version: "0.23.4"}, gomock.Any()).Do(func(_ interface{}, path string) error {
				os.MkdirAll(filepath.Join(path, "appdynamics-random-number", "bin"), 0755)
				ioutil.WriteFile(filepath.Join(path, "appdynamics-random-number", "bin", "binfile"), []byte("bin content"), 0755)
				os.MkdirAll(filepath.Join(path, "appdynamics-random-number", "lib"), 0755)
				ioutil.WriteFile(filepath.Join(path, "appdynamics-random-number", "lib", "libfile"), []byte("lib content"), 0644)
				return nil
			})
		})
		It("install desired version and links bin and lib directories", func() {
			Expect(supplier.InstallCrystal()).To(Succeed())
			Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "bin", "binfile"))).To(Equal([]byte("bin content")))
			Expect(ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "lib", "libfile"))).To(Equal([]byte("lib content")))
		})
	})

	Describe("InstallShards", func() {
		It("installs shards to cachedir", func() {
			mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) error {
				Expect(cmd.Dir).To(Equal(buildDir))
				Expect(cmd.Env).To(ContainElement(fmt.Sprintf("SHARDS_INSTALL_PATH=%s/shards_lib", cacheDir)))
				Expect(cmd.Args).To(Equal([]string{"shards", "install", "--production"}))
				return nil
			})
			Expect(supplier.InstallShards()).To(Succeed())
		})

		Context("appdynamics version < 0.25.0", func() {
			BeforeEach(func() {
				supplier.Shard.CrystalVersion = "0.24.2"
			})
			It("installs shards to cachedir", func() {
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) error {
					Expect(cmd.Dir).To(Equal(buildDir))
					Expect(cmd.Env).To(ContainElement(fmt.Sprintf("SHARDS_INSTALL_PATH=%s/shards_lib", cacheDir)))
					Expect(cmd.Args).To(Equal([]string{"appdynamics", "deps", "--production"}))
					return nil
				})
				Expect(supplier.InstallShards()).To(Succeed())
			})
		})
	})

	Describe("BuildApp", func() {
		Context("installed appdynamics has src in share/appdynamics/src", func() {
			BeforeEach(func() {
				supplier.Shard.Name = "appname"
				os.MkdirAll(filepath.Join(depsDir, depsIdx, "appdynamics-random-number", "share", "appdynamics", "src"), 0755)
			})
			It("builds the app using src from installed appdynamics and shards from cachedir", func() {
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) error {
					Expect(cmd.Dir).To(Equal(buildDir))
					Expect(cmd.Env).To(ContainElement(fmt.Sprintf("CRYSTAL_PATH=%s/%s/appdynamics-random-number/share/appdynamics/src:%s/shards_lib:src", depsDir, depsIdx, cacheDir)))
					Expect(cmd.Args).To(Equal([]string{
						"appdynamics", "build", "src/appname.cr", "--release",
						"-o", filepath.Join(depsDir, depsIdx, "app"),
					}))
					return nil
				})
				Expect(supplier.BuildApp()).To(Succeed())
			})
		})
		Context("installed appdynamics has src in src", func() {
			BeforeEach(func() {
				supplier.Shard.Name = "appname"
				os.MkdirAll(filepath.Join(depsDir, depsIdx, "appdynamics-random-number", "src"), 0755)
			})
			It("builds the app using src from installed appdynamics and shards from cachedir", func() {
				mockCommand.EXPECT().Run(gomock.Any()).Do(func(cmd *exec.Cmd) error {
					Expect(cmd.Dir).To(Equal(buildDir))
					Expect(cmd.Env).To(ContainElement(fmt.Sprintf("CRYSTAL_PATH=%s/%s/appdynamics-random-number/src:%s/shards_lib:src", depsDir, depsIdx, cacheDir)))
					Expect(cmd.Args).To(Equal([]string{
						"appdynamics", "build", "src/appname.cr", "--release",
						"-o", filepath.Join(depsDir, depsIdx, "app"),
					}))
					return nil
				})
				Expect(supplier.BuildApp()).To(Succeed())
			})
		})
	})
})
