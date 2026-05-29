package cli

import (
	gmmod "github.com/angei24/go-manager/internal/mod"
	"github.com/spf13/cobra"
)

var addUpgrade bool

var addCmd = &cobra.Command{
	Use:   "add <package>[@version] [packages...]",
	Short: "Add one or more module dependencies",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gmmod.Add(args, addUpgrade, verbose))
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <package> [packages...]",
	Short: "Remove one or more module dependencies",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exitErr(gmmod.Remove(args, verbose))
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
