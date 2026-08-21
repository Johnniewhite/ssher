//go:build windows

package pty

import "os/exec"

func preparePTYCommand(*exec.Cmd) {}
