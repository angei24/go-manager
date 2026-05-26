package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

// Execute runs the gm CLI.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "gm",
	Short: "Go toolchain manager (versions, modules, tools)",
	Long: `gm manages Go SDK versions, project dependencies, and global tools.
Inspired by uv for Python.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(goCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(toolCmd)
}

func exitErr(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
