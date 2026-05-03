package cmd

import "fmt"

func errNotImplemented(what string) error {
	return fmt.Errorf("%s is not yet implemented in the Go port", what)
}
