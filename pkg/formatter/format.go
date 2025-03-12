package formatter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/DataDog/gostackparse"
	"github.com/akerouanton/gostackdrill/pkg/grouping"
	"github.com/akerouanton/gostackdrill/pkg/utils"
)

type Options struct {
	Format      FormatType
	OmitStdlib  bool
	GroupBy     grouping.GroupByType
	Mod         string
	States      []string
	FrameFilter func(*gostackparse.Frame) bool
}

type FormatType string

const (
	FormatFull   FormatType = "full"
	FormatNoPath FormatType = "no-path"
	FormatShort  FormatType = "short"
	FormatStats  FormatType = "stats"
)

func PrintGoroutines(goroutines []*gostackparse.Goroutine, opts Options) {
	if opts.Format == FormatStats {
		PrintStats(goroutines, opts)
		return
	}

	slices.SortFunc(goroutines, func(a, b *gostackparse.Goroutine) int {
		return int(b.Wait - a.Wait)
	})

	var frameFilters []func(*gostackparse.Frame) bool
	if opts.OmitStdlib {
		frameFilters = append(frameFilters, utils.IsNotStdlibFrame)
	}
	if opts.FrameFilter != nil {
		frameFilters = append(frameFilters, opts.FrameFilter)
	}

	for _, goroutine := range goroutines {
		fmt.Println(formatGoroutine(goroutine, opts, frameFilters))
	}
}

func formatGoroutine(goroutine *gostackparse.Goroutine, opts Options, frameFilters []func(*gostackparse.Frame) bool) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("goroutine %d [%s", goroutine.ID, goroutine.State))
	if goroutine.Wait > 0 {
		b.WriteString(fmt.Sprintf(", %s", goroutine.Wait))
	}
	b.WriteString("]\n")

	// In non-full mode, only print the goroutine's state and wait time.
	if opts.Format == FormatShort {
		return b.String()
	}

	var ellision bool
	for _, frame := range goroutine.Stack {
		if !utils.FilterFrame(frame, frameFilters...) {
			ellision = true
			continue
		}

		if ellision {
			// b.WriteString("... (calls omitted) ...\n")
			ellision = false
		}

		b.WriteString(fmt.Sprintf("%s()\n", strings.TrimPrefix(frame.Func, opts.Mod+"/")))
		if opts.Format != FormatNoPath {
			b.WriteString(fmt.Sprintf("\t%s:%d\n", frame.File, frame.Line))
		}
	}

	if ellision {
		// b.WriteString("... (stdlib calls omitted) ...\n")
	}

	return b.String()
}

func PrintStats(goroutines []*gostackparse.Goroutine, opts Options) {
	groups := grouping.GroupGoroutines(goroutines, opts.GroupBy, opts.FrameFilter)

	for _, group := range groups {
		key := group.Key
		if opts.Mod != "" {
			key = strings.TrimPrefix(group.Key, opts.Mod+"/")
		}

		fmt.Printf("%s:\n\t%d goroutines, avg wait time: %s\n\n", key, len(group.Goroutines), group.AvgWaitTime())
	}
}
