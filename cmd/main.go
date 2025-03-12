package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/DataDog/gostackparse"
	"github.com/akerouanton/gostackdrill/pkg/formatter"
	"github.com/akerouanton/gostackdrill/pkg/grouping"
	"github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

type commonFlags struct {
	states     []string
	format     string
	omitStdlib bool
	groupBy    string
	filter     string
}

func (f commonFlags) toOptions() formatter.Options {
	var mod string
	modBytes, err := os.ReadFile("go.mod")
	if err == nil {
		mod = modfile.ModulePath(modBytes)
	}

	prefixFilter := f.filter
	if prefixFilter != "" && mod != "" && !govalidator.IsDNSName(prefixFilter) {
		prefixFilter = mod + "/" + prefixFilter
	}

	return formatter.Options{
		Format:     formatter.FormatType(f.format),
		OmitStdlib: f.omitStdlib,
		GroupBy:    grouping.GroupByType(f.groupBy),
		Mod:        mod,
		States:     f.states,
		FrameFilter: func(frame *gostackparse.Frame) bool {
			if prefixFilter != "" && frame != nil && !strings.HasPrefix(frame.Func, prefixFilter) {
				return false
			}
			return true
		},
	}
}

func addCommonFlags(cmd *cobra.Command, flags *commonFlags) {
	cmd.Flags().StringArrayVar(&flags.states, "states", []string{}, "states of goroutines to include")
	cmd.Flags().StringVar(&flags.format, "format", "short", "format of the output (short, full, stats)")
	cmd.Flags().BoolVar(&flags.omitStdlib, "omit-stdlib", false, "omit stdlib frames when --format=full")
	cmd.Flags().StringVar(&flags.groupBy, "group-by", "struct", "group by package, struct, or func")
	cmd.Flags().StringVar(&flags.filter, "filter", "", "filter goroutines by func prefix")
}

func main() {
	rootCmd := &cobra.Command{
		Use: "gostackdrill",
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(outliersCmd())
	rootCmd.AddCommand(printCmd())

	rootCmd.Execute()
}

func parseStackTrace(path string) ([]*gostackparse.Goroutine, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	defer f.Close()

	goroutines, errs := gostackparse.Parse(f)
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to parse file: %v", errors.Join(errs...))
	}

	return goroutines, nil
}
