// This file is used to detect build on unsupported GOOS/GOARCH combinations.

//go:build !linux && !darwin && !windows && !freebsd
//+build !linux,!darwin,!windows,!freebsd linux,!amd64,!arm64,!386,!arm darwin,!amd64 windows,!amd64 freebsd,!amd64

package your_operating_system_is_not_supported_by_delve
