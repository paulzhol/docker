// +build: freebsd

package jail

import (
	"fmt"
	"sync"
	"reflect"
	"syscall"
	"unsafe"
)

const (
	JAIL_CREATE int = 1 << iota /* Create jail if it doesn't exist */
	JAIL_UPDATE int = 1 << iota /* Update parameters of existing jail */
	JAIL_ATTACH int = 1 << iota /* Attach to jail upon creation */
	JAIL_DYING  int = 1 << iota /* Allow getting a dying jail */

	_JAIL_SET_MASK int = JAIL_CREATE | JAIL_UPDATE | JAIL_ATTACH | JAIL_DYING
	_JAIL_ERRMSGLEN = 1024
)

type JailError struct {
	Syscall string
	ErrMsg  string
	Err     error
}

func (e *JailError) Error() string {
	msg := e.Syscall + ": " + e.Err.Error()
	if len(e.ErrMsg) > 0 {
		msg += " (" + e.ErrMsg + ")"
	}
	return msg
}

type setLenFunc func(addr unsafe.Pointer, length int)

var (
	iovec_setlenimp_once sync.Once
	setLenImp setLenFunc
	zero_iovec syscall.Iovec
)

type jailParam struct {
	syscall.Iovec
}

func (iov *jailParam) SetLen(len int) {
	setLenImp(unsafe.Pointer(&iov.Len), len)
}

func (iov *jailParam) SetString(val string) (err error) {
	base, err := syscall.BytePtrFromString(val)
	if err != nil {
		return
	}
	iov.Base = base
	iov.SetLen(len(val)+1)
	return
}

func setLenUint32(addr unsafe.Pointer, len int) {
	*(*uint32)(addr) = uint32(len)
}

func setLenUint64(addr unsafe.Pointer, len int) {
	*(*uint64)(addr) = uint64(len)
}

func init() {
	iovec_setlenimp_once.Do(func() {
		v := reflect.ValueOf(zero_iovec.Len)
		switch kind := v.Kind(); kind {
		case reflect.Uint32:
			setLenImp = setLenUint32
		case reflect.Uint64:
			setLenImp = setLenUint64
		default:
			panic(fmt.Errorf("unsupported syscall.Iovec.Len Kind: %s", kind))
		}
	})
}

func Jail(jail_params map[string]interface{}, flags int) (jid int, err error) {
	iov := make([]jailParam, 2*(len(jail_params)+1))
	i := 0
	for param_name, param_val := range jail_params {
		val := reflect.ValueOf(param_val)
		switch kind := val.Kind(); kind {
		case reflect.Bool:
			if !val.Bool() {
				param_name = "no"+param_name
			}
			iov[i+1].Base = nil
			iov[i+1].SetLen(0)
		case reflect.String:
			iov[i+1].SetString(val.String())
		default:
			err = fmt.Errorf("unsupported jail param Kind: %s", kind)
			return
		}
		iov[i].SetString(param_name)
		i += 2
	}

	iov[i].SetString("errmsg")
	errmsg := make([]byte, _JAIL_ERRMSGLEN)
	iov[i+1].Base = &errmsg[0]
	iov[i+1].SetLen(len(errmsg))

	jid, err = jail_set(iov, flags)
	if err != nil {
		e := &JailError{Syscall: "jail_set", Err: err}
		for n, c := range errmsg {
			if c == 0 {
				e.ErrMsg = string(errmsg[0:n])
				break
			}
		}
		err = e
	}
	return
}

func jail_set(iov []jailParam, flags int) (int, error) {
	r0, _, e1 := syscall.Syscall(syscall.SYS_JAIL_SET, uintptr(unsafe.Pointer(&iov[0])), uintptr(len(iov)), uintptr(flags))
	jid := int(r0)
	if e1 != 0 {
		return jid, e1
	}
	return jid, nil
}
