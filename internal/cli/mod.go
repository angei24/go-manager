package cli

import (
	gmmod "github.com/angei24/go-manager/internal/mod"
	"github.com/spf13/cobra"
)

var addUpgrade bool

var addCmd = &cobra.Command{
	Use:   "add <package>[@version]",
	Short: "Add a module dependency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gmmod.Add(args[0], addUpgrade, verbose))
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Remove a module dependency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gmmod.Remove(args[0], verbose))
	},
}

var syncCheck bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Tidy and download module dependencies (go mod tidy + download)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gmmod.Sync(syncCheck, verbose))
	},
}

func init() {
	addCmd.Flags().BoolVarP(&addUpgrade, "upgrade", "u", false, "upgrade existing dependency")
	syncCmd.Flags().BoolVar(&syncCheck, "check", false, "exit non-zero if go.mod/go.sum would change")
}
