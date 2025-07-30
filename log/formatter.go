package log

import (
	"runtime"
	"strings"
)

const (
	maxCallersStackGet = 6
	skippedCallers     = 4
	unknownCaller      = "unknown"
	pathSegmentLimit   = 4
)

func getCaller() (string, string) {
	pcs := make([]uintptr, maxCallersStackGet)
	n := runtime.Callers(skippedCallers, pcs)
	if n == 0 {
		return unknownCaller, unknownCaller
	}
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)

	for {
		frame, more := frames.Next()
		if frame.File != "" && frame.Function != "" {
			file := _shortenPath(frame.File, pathSegmentLimit)
			function := _shortenFunction(frame.Function, pathSegmentLimit)
			return file, function
		}
		if !more {
			break
		}
	}

	return unknownCaller, unknownCaller
}

func _shortenPath(path string, limit int) string {
	segments := strings.Split(path, "/")
	if len(segments) <= limit {
		return path
	}
	return strings.Join(segments[len(segments)-limit:], "/")
}

func _shortenFunction(function string, limit int) string {
	segments := strings.Split(function, "/")
	if len(segments) <= limit {
		return function
	}
	return strings.Join(segments[len(segments)-limit:], "/")
}
