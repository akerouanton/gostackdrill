package main

import (
	"log"

	"github.com/DataDog/gostackparse"
	"github.com/akerouanton/gostackdrill/pkg/formatter"
	"github.com/akerouanton/gostackdrill/pkg/utils"
	"github.com/spf13/cobra"
)

type printFlags struct {
	commonFlags
	id int
}

func printCmd() *cobra.Command {
	var flags printFlags

	cmd := &cobra.Command{
		Use: "print",
		Run: func(cmd *cobra.Command, args []string) {
			goroutines, err := parseStackTrace(args[0])
			if err != nil {
				log.Fatalf("failed to parse file: %v", err)
			}

			formatterOpts := flags.commonFlags.toOptions()

			var filtered []*gostackparse.Goroutine
			for _, goroutine := range goroutines {
				if flags.id != 0 && goroutine.ID != flags.id {
					// log.Printf("goroutine %d != filter %d\n", goroutine.ID, flags.id)
					continue
				}

				frame, _ := utils.FindFirstFrame(goroutine.Stack, formatterOpts.FrameFilter)
				if frame == nil {
					log.Printf("goroutine %d has no frame matching filter criteria\n", goroutine.ID)
					continue
				}

				filtered = append(filtered, goroutine)
			}

			formatterOpts.FrameFilter = nil
			formatter.PrintGoroutines(filtered, formatterOpts)
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().IntVar(&flags.id, "id", 0, "id of the goroutine to print")
	addCommonFlags(cmd, &flags.commonFlags)

	return cmd
}
