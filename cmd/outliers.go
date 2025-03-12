package main

import (
	"log"
	"slices"
	"strings"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/akerouanton/gostackdrill/pkg/formatter"
	"github.com/spf13/cobra"
)

type outliersFlags struct {
	commonFlags
	threshold time.Duration
}

var defaultStates = []string{
	"syscall",
	"sync.Cond.Wait",
	"sync.Mutex.Lock",
}

func outliersCmd() *cobra.Command {
	var flags outliersFlags

	cmd := &cobra.Command{
		Use: "outliers [flags] <file>",
		Run: func(cmd *cobra.Command, args []string) {
			flags.states = mergeStates(flags.states, defaultStates)
			runOutliers(args[0], flags)
		},
		Args:  cobra.ExactArgs(1),
		Short: "Filter goroutines with a symptomatic wait time",
		Long: `Filter goroutines with a symptomatic wait time.

States can be prefixed with a '+' or '-' to include or exclude them from the default list.
Default states include: ` + strings.Join(defaultStates, ", ") + `.

The default threshold used to determine if a goroutine is an outlier is ` + flags.threshold.String() + `.
It can be overridden with the --threshold flag.

Supported formats:
- short: print the goroutine id and wait time
- full: print the goroutine id, wait time, and stack trace
- stats: print the number of outliers and the average wait time.

When --format=stats, goroutines can be grouped by:
- package: group by the package of the first non-stdlib frame
- struct: group by the struct of the first non-stdlib frame
- func: group by the function of the first non-stdlib frame
`,
	}

	cmd.Flags().DurationVar(&flags.threshold, "threshold", 10*time.Second, "threshold for outliers")
	addCommonFlags(cmd, &flags.commonFlags)

	return cmd
}

func mergeStates(states, defaultStates []string) []string {
	var finalStates []string
	defaults := defaultStates
	includeDefaults := true

	for _, state := range states {
		if state == "-all" {
			includeDefaults = false
			continue
		}

		if len(state) > 0 && state[0] == '-' {
			state = state[1:]
			defaults = slices.DeleteFunc(defaults, func(s string) bool {
				return s == state
			})
		}

		if len(state) > 0 && state[0] == '+' {
			finalStates = append(finalStates, state)
		}
	}

	if includeDefaults {
		finalStates = append(finalStates, defaults...)
	}

	return finalStates
}

func runOutliers(path string, flags outliersFlags) {
	formatterOpts := flags.commonFlags.toOptions()

	goroutines, err := parseStackTrace(path)
	if err != nil {
		log.Fatalf("failed to parse file: %v", err)
	}

	var outliers []*gostackparse.Goroutine
	for _, goroutine := range goroutines {
		if goroutine.Wait < flags.threshold {
			continue
		}

		outliers = append(outliers, goroutine)
	}

	formatter.PrintGoroutines(outliers, formatterOpts)
}
