package cli

import (
	"github.com/angei24/go-manager/internal/tool"
	"github.com/spf13/cobra"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Manage globally installed Go tools",
}

var toolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(tool.List())
	},
}

var toolInstallCmd = &cobra.Command{
	Use:   "install <package>[@version]",
	Short: "Install a Go tool binary",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(tool.Install(args[0], verbose))
	},
}

var toolUninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Uninstall a Go tool by binary name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(tool.Uninstall(args[0]))
	},
}

func init() {
	toolCmd.AddCommand(toolListCmd, toolInstallCmd, toolUninstallCmd)
}
