package cli

import (
	"github.com/angei24/go-manager/internal/project"
	"github.com/spf13/cobra"
)

var (
	initModule string
	initNoGit  bool
	initForce  bool
)

var initCmd = &cobra.Command{
	Use:   "init [dir]",
	Short: "Create a new Go project with git, go.mod, and templates",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		return exitErr(project.Init(project.InitOptions{
			Dir:     dir,
			Module:  initModule,
			NoGit:   initNoGit,
			Force:   initForce,
			Verbose: verbose,
		}))
	},
}

func init() {
	initCmd.Flags().StringVar(&initModule, "module", "", "module path (default: <dirname>)")
	initCmd.Flags().BoolVar(&initNoGit, "no-git", false, "skip git init")
	initCmd.Flags().BoolVar(&initForce, "force", false, "allow non-empty directory")
}
