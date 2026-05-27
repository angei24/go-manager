package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angei24/go-manager/internal/gover"
	"github.com/angei24/go-manager/internal/mod"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Manage Go SDK versions",
}

var goListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Go versions and latest supported stable releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gover.List(verbose))
	},
}

var goInstallCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Download and install a Go version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gover.Install(args[0], verbose))
	},
}

var goUseGlobal bool

var goUseCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Set Go version (.gm-version and go.mod in project)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		display, err := gover.Use(args[0], goUseGlobal)
		if err != nil {
			return exitErr(err)
		}
		if goUseGlobal {
			return nil
		}
		wd, err := os.Getwd()
		if err != nil {
			return exitErr(err)
		}
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err != nil {
			fmt.Printf("Project Go version set to %s (.gm-version)\n", display)
			return nil
		}
		if err := mod.PinProjectGoVersion(wd, display); err != nil {
			return exitErr(err)
		}
		fmt.Printf("Project Go version set to %s (.gm-version and go.mod)\n", display)
		return nil
	},
}

var goUninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Remove an installed Go version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gover.Uninstall(args[0]))
	},
}

func init() {
	goUseCmd.Flags().BoolVar(&goUseGlobal, "global", false, "set global default in config")
	goCmd.AddCommand(goListCmd, goInstallCmd, goUseCmd, goUninstallCmd)
}
