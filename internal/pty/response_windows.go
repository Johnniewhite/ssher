//go:build windows

package pty

// ConPTY consumes virtual-terminal input. Carriage return is the Enter key;
// a bare line feed is output control and is not guaranteed to submit a prompt.
func formatPromptResponse(password string) string { return password + "\r" }
