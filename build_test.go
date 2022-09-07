package buildpkg

import (
	"log"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestBuild(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test on CI")
	}
	c := qt.New(t)

	opts := Options{
		Infof: func(format string, args ...interface{}) {
			log.Printf(format, args...)
		},
		Dir:                  "./testdata",
		SigningIdentity:      "ZYSJUFSYL4",
		StagingDirectory:     "./staging",
		Identifier:           "is.bep.helloworld",
		Version:              "0.0.13",
		InstallLocation:      "/usr/local/bin",
		PackageOutputPath:    "./helloworld.pkg",
		SkipCodeSigning:      false,
		SkipNotarization:     false,
		SkipInstallerSigning: false,
		//ScriptsDirectory: "./testdata/scripts",
	}

	builder, err := New(opts)
	c.Assert(err, qt.IsNil)
	c.Assert(builder.runCommand("./prepare.sh", builder.Version), qt.IsNil)
	err = builder.Build()
	c.Assert(err, qt.IsNil)

}
