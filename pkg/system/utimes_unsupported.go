// +build !linux,!dragonfly,!freebsd,!netbsd

package system

import "syscall"

func LUtimesNano(path string, ts []syscall.Timespec) error {
	return ErrNotSupportedPlatform
}
