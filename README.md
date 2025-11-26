[![Tests on Linux, MacOS and Windows](https://github.com/bep/buildpkg/workflows/Test/badge.svg)](https://github.com/bep/buildpkg/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/buildpkg)](https://goreportcard.com/report/github.com/bep/buildpkg)
[![GoDoc](https://godoc.org/github.com/bep/buildpkg?status.svg)](https://godoc.org/github.com/bep/buildpkg)

This journey started with my naive idea that I could do all of this outside of Macintosh/MacOS. It started out great when I found Apple's [Notary API](https://developer.apple.com/documentation/notaryapi), so I wrote [macosnotarylib](https://github.com/bep/macosnotarylib). But I also needed binary and package signing, and I gave up the grand idea. There are some third party libraries that claim to do some of this, but for me, even getting passed the trust part where I would have to inspect the code, would be too much work.

So, I wrote some tooling for this myself that uses Apple's CLI tools and API, and this library is the core part of it. There are Go alternatives out there, most notably [gon](https://github.com/mitchellh/gon). The biggest difference is that `gon` produces DMG or ZIP files. This library produces the very end-user-friendly PKG format.

The bulding blocks:

* Running `buildpkg.New(opts).Build()` will
    1. Sign the binary with `codesign`
    1. Package the binary with `pkgbuild`
    1. Sign the package with `productsign`
    1. Check the package with `pkgutil`
    1. Notarize the package with [macosnotarylib](https://github.com/bep/macosnotarylib) (uses the Apple API)
    1. Staple the package with `stapler`

For the **codesign** step you need create a `Developer ID Application Certificate` and for the **package signing** step you need a `Developer ID Installer Certificate`. These needs to be imported into your Keychain. Follow the instructions at [[developer.apple.com](https://developer.apple.com/account/resources/certificates/list](https://developer.apple.com/help/account/certificates/create-developer-id-certificates/)).

<img width="1028" alt="image" src="https://user-images.githubusercontent.com/394382/189410218-cab4cbf9-4f82-4f4b-ab0a-f19eb90e9c20.png">

Once you have those imported in the Keychain you can locate their common _signing identity_ with `security find-identity -v`, which is `XYZJUFSYL4` in the example below:

```bash
~/d/g/hugoreleaser ❯❯❯ security find-identity -v
  1) D4A412805301423E2DF63D90CE37C8A050B3AA2F "Developer ID Application: Bjørn Erik Pedersen (XYZJUFSYL4)"
  2) D4A412805301423E2DF63D90CE37C8A050B3AA2F "Developer ID Application: Bjørn Erik Pedersen (XYZJUFSYL4)"
  3) EADAD38B73CADB2E6975F55B8735F17B09138217 "Developer ID Installer: Bjørn Erik Pedersen (XYZJUFSYL4)"
  4) EADAD38B73CADB2E6975F55B8735F17B09138217 "Developer ID Installer: Bjørn Erik Pedersen (XYZJUFSYL4)"
     4 valid identities found
```

For the **notarizer** step you need to [create a new new API access key](https://appstoreconnect.apple.com/access/api) with `Developer` access and download the private key. Take note of the `Issuer ID` and `Key ID`:

<img width="1025" alt="image" src="https://user-images.githubusercontent.com/394382/189411457-d0ecf2f8-5457-45ad-ae0c-bd48fd48ab5a.png">

Also See [Creating API Keys for App Store Connect AP](https://developer.apple.com/documentation/appstoreconnectapi/creating_api_keys_for_app_store_connect_api).

With the above you could put the _signing identity_ in the `Options` struct and pass it to `New`:

```go

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

	// E.g. /usr/local/bin
	InstallLocation string

	// Scripts passed on the command line --scripts flag.
	// E.g. /mypkgscripts
	ScriptsDirectory string

	// Flags to enable skipping of build steps.
	SkipCodeSigning      bool
	SkipInstallerSigning bool
	SkipNotarization     bool
}
```

The other settings currently needs to be set as OS environment variables:

*  `MACOSNOTARYLIB_ISSUER_ID`
*  `MACOSNOTARYLIB_KID` (Key ID)
*  `MACOSNOTARYLIB_PRIVATE_KEY` (in base64 format).

## Use with Hugoreleaser

There are 2 archive plugins available:

* [macospkgremote](https://github.com/gohugoio/hugoreleaser-archive-plugins/tree/main/macospkgremote)
* [macospkg](https://github.com/gohugoio/hugoreleaser-archive-plugins/tree/main/macospkg) (a "local" variant of the above)

Also see it configured in Hugoreleaser's [build config](https://github.com/gohugoio/hugoreleaser/blob/main/hugoreleaser.toml).
