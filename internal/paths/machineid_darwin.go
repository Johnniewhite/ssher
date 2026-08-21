//go:build darwin

package paths

import (
	"fmt"
	"os/exec"
	"strings"
)

func readMachineID() (string, error) {
	out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "", err
	}
	const marker = "IOPlatformUUID"
	idx := strings.Index(string(out), marker)
	if idx < 0 {
		return "", fmt.Errorf("IOPlatformUUID not found")
	}
	segment := string(out)[idx+len(marker):]
	firstQuote := strings.Index(segment, `"`)
	if firstQuote < 0 {
		return "", fmt.Errorf("malformed ioreg output")
	}
	segment = segment[firstQuote+1:]
	secondQuote := strings.Index(segment, `"`)
	if secondQuote < 0 {
		return "", fmt.Errorf("malformed ioreg output")
	}
	return segment[:secondQuote], nil
}
