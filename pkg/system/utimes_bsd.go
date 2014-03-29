// +build dragonfly freebsd netbsd

package system

import (
	"syscall"
	"unsafe"
)

func LUtimesNano(path string, ts []syscall.Timespec) error {
	if len(ts) != 2 {
			return syscall.EINVAL
	}
	// Not as efficient as it could be because Timespec and
	// Timeval have different types in the different OSes
	tv := [2]syscall.Timeval{
			syscall.NsecToTimeval(syscall.TimespecToNsec(ts[0])),
			syscall.NsecToTimeval(syscall.TimespecToNsec(ts[1])),
	}
	var _p0 *byte
	_p0, err := syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	_, _, e1 := syscall.Syscall(syscall.SYS_LUTIMES, uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(&tv)), 0)
	if e1 != 0 {
		return e1
	}
	return nil
}
