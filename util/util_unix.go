//go:build !windows
// +build !windows

package util

func isPageantAvailable() bool {
	return false
}
