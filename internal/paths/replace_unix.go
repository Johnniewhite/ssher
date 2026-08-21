//go:build !windows

package paths

import "os"

func replaceFile(source, target string) error { return os.Rename(source, target) }
