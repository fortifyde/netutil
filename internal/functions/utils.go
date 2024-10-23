package functions

import (
	"os"
	"os/exec"
)

// check dependencies to ensure required binaries are installed
func CheckDependencies(binaries []string) bool {
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin); err != nil {
			return false
		}
	}
	return true
}

// ensure directory exists and create if it does not
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}
	return nil
}
