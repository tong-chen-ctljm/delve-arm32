// This file is used to detect build on unsupported GOOS/GOARCH combinations.

//go:build linux && !amd64 && !arm64 && !386 && !arm
// +build linux,!amd64,!arm64,!386,!arm

package your_linux_architecture_is_not_supported_by_delve
