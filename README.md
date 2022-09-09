[![Tests on Linux, MacOS and Windows](https://github.com/bep/buildpkg/workflows/Test/badge.svg)](https://github.com/bep/buildpkg/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/buildpkg)](https://goreportcard.com/report/github.com/bep/buildpkg)
[![GoDoc](https://godoc.org/github.com/bep/buildpkg?status.svg)](https://godoc.org/github.com/bep/buildpkg)

This journey started with my naive idea that I could do all of this outside of Macintosh/MacOS. It started out great when I found Apple's [Notary API](https://developer.apple.com/documentation/notaryapi), so I wrote [macosnotarylib](https://github.com/bep/macosnotarylib). But I would also at least have code and package signing place, and this was where I had to give up the grand. There are some third party libraries that claim to do some of this, but for me, even getting passed the trust part where I would have to inspect the code, would be too much work.

So, I wrote some tooling for this myself that uses Apple's CLI tools and API, and this library is the core part of it. There are Go alternatives out there, most notably [gon](https://github.com/mitchellh/gon). The biggest difference is that `gon` produces DMG or ZIP files. This library produces the very end-user-friendly PKG format.

The bulding blocks:

* Running `buildpkg.New(opts).Build()` will
    1. Sign the binary with `codesign`
    1. Package the binary with `pkgbuild`
    1. Sign the package with `productsign`
    1. Check the package with `pkgutil`
    1. Notarize the package with [macosnotarylib](https://github.com/bep/macosnotarylib) (uses the Apple API)
    1. Staple the package with `stapler`
