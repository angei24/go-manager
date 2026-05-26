package cli

import (
	"github.com/spf13/cobra"
	"github.com/angei24/go-manager/internal/gover"
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
	Short: "Set default Go version (project .gm-version or --global)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gover.Use(args[0], goUseGlobal))
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
