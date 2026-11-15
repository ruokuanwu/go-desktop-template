package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go-desktop-template/internal/app"
)

func main() {
	command := &cobra.Command{
		Use:           "todo-desktop",
		Short:         "Todo Desktop",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := os.LookupEnv("MYGO_GPU"); !ok {
				if err := os.Setenv("MYGO_GPU", "0"); err != nil {
					return fmt.Errorf("set default renderer: %w", err)
				}
			}
			return app.Run()
		},
	}
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
