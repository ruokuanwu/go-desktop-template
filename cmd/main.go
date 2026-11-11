package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"go-desktop-template/internal/app"
)

func main() {
	var configPath string
	command := &cobra.Command{
		Use:           "todo-desktop",
		Short:         "Todo Desktop",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				base, err := os.UserConfigDir()
				if err != nil {
					return err
				}
				configPath = filepath.Join(base, "go-desktop-template", "config.json")
			}
			path, err := filepath.Abs(configPath)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("config-path") {
				if _, err := os.Stat(path); err != nil {
					return fmt.Errorf("config path: %w", err)
				}
			}
			return app.Run(path)
		},
	}
	command.Flags().StringVarP(&configPath, "config-path", "p", "", "配置文件路径（默认为系统用户配置目录/go-desktop-template/config.json）")
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
