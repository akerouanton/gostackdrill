package utils

import (
	"strings"

	"github.com/DataDog/gostackparse"
)

func IsStdlibFrame(frame *gostackparse.Frame) bool {
	return strings.HasPrefix(frame.File, "runtime/") ||
		strings.HasPrefix(frame.File, "internal/") ||
		strings.HasPrefix(frame.File, "net/") ||
		strings.HasPrefix(frame.File, "sync/") ||
		strings.HasPrefix(frame.File, "os/") ||
		strings.HasPrefix(frame.File, "syscall/") ||
		strings.HasPrefix(frame.File, "io/") ||
		strings.HasPrefix(frame.File, "bufio/") ||
		strings.HasPrefix(frame.File, "encoding/")
}

func IsNotStdlibFrame(frame *gostackparse.Frame) bool {
	return !IsStdlibFrame(frame)
}

func FindFirstFrame(stack []*gostackparse.Frame, filters ...func(*gostackparse.Frame) bool) (*gostackparse.Frame, bool) {
	var ellision bool
	for _, frame := range stack {
		if !FilterFrame(frame, filters...) {
			ellision = true
			continue
		}
		return frame, ellision
	}
	return nil, false
}

func FilterFrame(frame *gostackparse.Frame, filters ...func(*gostackparse.Frame) bool) bool {
	for _, filter := range filters {
		if !filter(frame) {
			return false
		}
	}
	return true
}
