//go:build windows
// +build windows

package util

import "github.com/davidmz/go-pageant"

func isPageantAvailable() bool {
	return pageant.Available()
}
