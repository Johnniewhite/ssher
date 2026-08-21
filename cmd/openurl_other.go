//go:build !darwin && !linux && !windows

package cmd

import "fmt"

func openURL(string) error { return fmt.Errorf("opening a browser is not supported on this platform") }
