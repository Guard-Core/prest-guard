package main

// main exists so the module also builds as an ordinary binary (go build ./...)
// for tooling and CI. The artifact that matters is the plugin:
//
//	go build -buildmode=plugin -trimpath -ldflags "-s -w" -o guard.so .
//
// See the README for the toolchain and dependency-version constraints that
// native Go plugins carry.
func main() {}
