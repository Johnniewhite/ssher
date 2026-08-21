//go:build darwin

package cmd

import "os/exec"

func openURL(url string) error { return exec.Command("open", url).Start() }
