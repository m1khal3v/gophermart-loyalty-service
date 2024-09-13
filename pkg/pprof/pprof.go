// Package pprof
// contains facade for runtime/pprof package
package pprof

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

type profile string

const (
	Heap         profile = "heap"
	Goroutine    profile = "goroutine"
	Allocs       profile = "allocs"
	ThreadCreate profile = "threadcreate"
	Block        profile = "block"
	Mutex        profile = "mutex"
)

// CPUCapture creates file and start CPU profiling for specified duration to it
func CPUCapture(ctx context.Context, filename string, duration time.Duration) error {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := pprof.StartCPUProfile(file); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-time.After(duration):
		// capture finished
	}

	pprof.StopCPUProfile()

	return file.Close()
}

// Capture creates file and save specified profiling to it
func Capture(profile profile, filename string) error {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	pprofile := pprof.Lookup(string(profile))
	if pprofile == nil {
		return fmt.Errorf("pprof profile not found")
	}

	if err := pprofile.WriteTo(file, 0); err != nil {
		return err
	}

	return file.Close()
}
