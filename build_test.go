package buildpkg

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestBuild(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test on CI")
	}
	c := qt.New(t)

	wd, err := os.Getwd()
	c.Assert(err, qt.IsNil)
	testData := filepath.Join(wd, "testdata")
	opts := Options{
		Infof: func(format string, args ...interface{}) {
			log.Printf(format, args...)
		},
		Dir:                   testData,
		SigningIdentity:       "ZYSJUFSYL4",
		SigningEntitlements:   []string{"com.apple.security.cs.allow-jit", "com.apple.security.cs.allow-unsigned-executable-memory"},
		StagingDirectory:      filepath.Join(testData, "staging"),
		Identifier:            "is.bep.helloworld",
		Version:               "0.0.13",
		InstallLocation:       "/usr/local/bin",
		PackageOutputFilename: filepath.Join(testData, "helloworld.pkg"),
		SkipCodeSigning:       false,
		SkipNotarization:      false,
		SkipInstallerSigning:  false,
	}

	builder, err := New(opts)
	c.Assert(err, qt.IsNil)
	c.Assert(builder.runCommand("./prepare.sh", builder.Version), qt.IsNil)
	err = builder.Build()
	c.Assert(err, qt.IsNil)
}
