package buildpkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bep/macosnotarylib"
	"github.com/golang-jwt/jwt/v4"
)

// New creates a new Builder.
func New(opts Options) (*Builder, error) {
	if err := opts.init(); err != nil {
		return nil, err
	}
	return &Builder{Options: opts}, nil
}

type Builder struct {
	Options
}

// Build signs the binary, builds the package and signs and notarizes and staples it.
// It' currently limited to 1 file only.
// The notarization part requires the following environment variables to be set:
// - MACOSNOTARYLIB_ISSUER_ID
// - MACOSNOTARYLIB_KID
// - MACOSNOTARYLIB_PRIVATE_KEY (in base64 format).
func (b *Builder) Build() error {
	files, err := os.ReadDir(b.StagingDirectory)
	if err != nil {
		return err
	}
	if len(files) != 1 || files[0].IsDir() {
		return fmt.Errorf("opts: StagingDirectory must contain exactly one file")
	}

	if !b.SkipCodeSigning {
		for _, fi := range files {
			if err := b.runCommand("codesign", "-s", b.SigningIdentity, "--options=runtime", filepath.Join(b.StagingDirectory, fi.Name())); err != nil {
				return err
			}
		}
	}

	tempPackageOutputFilename := b.PackageOutputFilename + ".tmp"

	args := []string{
		"--root", b.StagingDirectory,
		"--identifier", b.Identifier,
		"--version", b.Version,
		"--install-location", b.InstallLocation,
	}

	if b.ScriptsDirectory != "" {
		args = append(args, "--scripts", b.ScriptsDirectory)
	}

	if b.SkipInstallerSigning {
		tempPackageOutputFilename = b.PackageOutputFilename
	}

	args = append(args, tempPackageOutputFilename)

	if err := b.runCommand("pkgbuild", args...); err != nil {
		return err
	}

	if !b.SkipInstallerSigning {
		// Sign the package
		if err := b.runCommand("productsign", "--sign", b.SigningIdentity, tempPackageOutputFilename, b.PackageOutputFilename); err != nil {
			return err
		}

		if err := os.Remove(tempPackageOutputFilename); err != nil {
			return err
		}

		// Check the package signature.
		if err := b.runCommand("pkgutil", "--check-signature", b.PackageOutputFilename); err != nil {
			return err
		}
	}

	if !b.SkipNotarization {
		// Notarize the package.
		if err := b.notarizePackage(filepath.Join(b.PackageOutputFilename)); err != nil {
			return err
		}

		// Staple the package.
		if err := b.runCommand("stapler", "staple", b.PackageOutputFilename); err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) notarizePackage(filename string) error {
	issuerID := os.Getenv("MACOSNOTARYLIB_ISSUER_ID")

	kid := os.Getenv("MACOSNOTARYLIB_KID")
	if kid == "" || issuerID == "" {
		return fmt.Errorf("env: MACOSNOTARYLIB_ISSUER_ID and MACOSNOTARYLIB_KID must be set")
	}

	// This test also depends on the private key from env MACOSNOTARYLIB_PRIVATE_KEY in base64 format. See below.

	n, err := macosnotarylib.New(
		macosnotarylib.Options{
			InfoLoggerf: b.Infof,
			IssuerID:    issuerID,
			Kid:         kid,
			SignFunc: func(token *jwt.Token) (string, error) {
				key, err := macosnotarylib.LoadPrivateKeyFromEnvBase64("MACOSNOTARYLIB_PRIVATE_KEY")
				if err != nil {
					return "", err
				}
				return token.SignedString(key)
			},
		},
	)

	if err != nil {
		return err
	}

	return n.Submit(filename)

}

func (b *Builder) runCommand(name string, args ...string) error {
	fmt.Println(name, args)
	cmd := exec.Command(name, args...)
	cmd.Dir = b.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Options struct {
	// The Info logger.
	// If nil, no Info logging will be done.
	Infof func(format string, a ...interface{})

	// The Dir to build from.
	Dir string

	// Developer ID Application + Developer ID Installer
	// https://developer.apple.com/account/resources/certificates/list
	SigningIdentity string

	// The result
	PackageOutputFilename string

	// The staging directory where all your build artifacts are located.
	StagingDirectory string

	// E.g. io.gohugo.hugo
	Identifier string

	// E.g. 234
	Version string

	// E.g. /Applications
	InstallLocation string

	// Scripts passed on the command line --scripts flag.
	// E.g. /mypkgscripts
	ScriptsDirectory string

	// Flags to enable skipping of build steps.
	SkipCodeSigning      bool
	SkipInstallerSigning bool
	SkipNotarization     bool
}

func (o *Options) init() error {
	if o.Infof == nil {
		o.Infof = func(format string, a ...interface{}) {
		}
	}
	if o.SigningIdentity == "" {
		return fmt.Errorf("opts: SigningIdentity is required")
	}

	if o.StagingDirectory == "" {
		return fmt.Errorf("opts: StagingDirectory not set")
	}

	if o.Identifier == "" {
		return fmt.Errorf("opts: Identifier not set")
	}

	if o.Version == "" {
		return fmt.Errorf("opts: Version not set")
	}

	if o.InstallLocation == "" {
		return fmt.Errorf("opts: InstallLocation not set")
	}

	if o.PackageOutputFilename == "" {
		return fmt.Errorf("opts: PackageOutputPath not set")
	}

	return nil
}
