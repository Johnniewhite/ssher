//go:build !darwin && !linux && !windows

package paths

import (
	"fmt"
	"runtime"
)

func readMachineID() (string, error) {
	return "", fmt.Errorf("unsupported os: %s", runtime.GOOS)
}
