// +build darwin dragonfly freebsd netbsd openbsd

package vfs

import (
	"fmt"
	"os/exec"
)

func copyDir(src, dst string) error {
	if output, err := exec.Command("cp", "-a", src+"/", dst).CombinedOutput(); err != nil {
		return fmt.Errorf("Error VFS copying directory: %s (%s)", err, output)
	}
	return nil
}
