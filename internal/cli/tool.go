package cli

import (
	"fmt"

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
	Use:   "install <package>[@version] [packages...]",
	Short: "Install one or more Go tool binaries",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, pkg := range args {
			if err := tool.Install(pkg, verbose); err != nil {
				return exitErr(fmt.Errorf("%s: %w", pkg, err))
			}
		}
		return nil
	},
}

var toolUninstallCmd = &cobra.Command{
	Use:   "uninstall <name> [names...]",
	Short: "Uninstall one or more Go tools by binary name",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, name := range args {
			if err := tool.Uninstall(name); err != nil {
				return exitErr(fmt.Errorf("%s: %w", name, err))
			}
		}
		return nil
	},
}

func init() {
	toolCmd.AddCommand(toolListCmd, toolInstallCmd, toolUninstallCmd)
}
