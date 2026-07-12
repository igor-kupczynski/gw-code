package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/igor-kupczynski/gw-code/internal/app"
	"github.com/igor-kupczynski/gw-code/internal/buildinfo"
	"github.com/igor-kupczynski/gw-code/internal/config"
	"github.com/igor-kupczynski/gw-code/internal/groveapi"
	"github.com/spf13/cobra"
)

const exitError = 1

// Run executes the gw-code CLI.
func Run() int {
	return runWithArgs(os.Args[1:])
}

func runWithArgs(args []string) int {
	env, err := groveapi.EnvFromOS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitError
	}
	client := groveapi.NewExecClient(env)
	resolver := groveapi.NewResolver(client, env)
	application := &app.App{
		Resolver:     resolver,
		GlobalConfig: config.GlobalPath(env.Dir),
		Stderr:       os.Stderr,
		LookPath:     exec.LookPath,
	}

	root := buildRoot(application)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitError
	}
	return 0
}

func buildRoot(application *app.App) *cobra.Command {
	var refreshFlag bool
	var pathFlag bool

	root := &cobra.Command{
		Use:   "gw-code [WORKSPACE]",
		Short: "Open Grove workspaces in VS Code or Cursor",
		Long:  "Generate multi-root editor workspaces for Grove and launch VS Code or Cursor in a new window.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace := workspaceArg(args)
			ctx := context.Background()
			switch {
			case pathFlag:
				path, err := application.Path(ctx, workspace)
				if err != nil {
					return err
				}
				fmt.Println(path)
				return nil
			case refreshFlag:
				result, err := application.Refresh(ctx, workspace)
				if err != nil {
					return err
				}
				fmt.Println(result.WorkspaceFilePath)
				if result.Changed {
					fmt.Fprintln(os.Stderr, "refreshed", result.WorkspaceFilePath)
				} else {
					fmt.Fprintln(os.Stderr, "unchanged", result.WorkspaceFilePath)
				}
				return nil
			default:
				path, err := application.Open(ctx, workspace)
				if err != nil {
					return err
				}
				fmt.Println(path)
				return nil
			}
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.Version = buildinfo.Version
	root.SetVersionTemplate("gw-code {{.Version}}\n")
	root.SilenceUsage = true
	root.Flags().BoolVar(&refreshFlag, "refresh", false, "Generate workspace file and print its path without opening the editor")
	root.Flags().BoolVar(&pathFlag, "path", false, "Print the deterministic workspace file path without generating or opening")
	root.MarkFlagsMutuallyExclusive("refresh", "path")
	return root
}

func workspaceArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
