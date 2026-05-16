package runtimeutils

import (
	"log/slog"
	"runtime"
	"strconv"
)

// Stack represents a stack of program counters of function invocations.
type Stack []uintptr

func CaptureStack(skip int) *Stack {
	const depth = 32

	var pcs [depth]uintptr

	n := runtime.Callers(skip+1, pcs[:])

	var st Stack = pcs[0:n]

	return &st
}

func (s *Stack) CallersFrames() []runtime.Frame {
	frames := runtime.CallersFrames(*s)

	var result []runtime.Frame

	for {
		frame, more := frames.Next()
		result = append(result, frame)

		if !more {
			break
		}
	}

	return result
}

func (s *Stack) StringLines() []string {
	frames := s.CallersFrames()

	result := make([]string, 0, len(frames))
	for _, frame := range frames {
		result = append(result,
			frame.File+":"+strconv.FormatInt(int64(frame.Line), 10),
		)
	}

	return result
}

func (s *Stack) SlogLines() slog.Attr {
	return slog.Any("stack", s.StringLines())
}
