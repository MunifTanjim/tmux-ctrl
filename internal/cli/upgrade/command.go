package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/MunifTanjim/tmux-ctrl/internal/cli/completion"
	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := util.EnsureTool("gh"); err != nil {
				return err
			}

			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("failed to get executable path: %w", err)
			}
			if execPath, err = filepath.EvalSymlinks(execPath); err != nil {
				return fmt.Errorf("failed to resolve executable path: %w", err)
			}
			if filepath.Base(execPath) != config.BinaryName {
				return fmt.Errorf("upgrade can only be run with the %s binary", config.BinaryName)
			}

			// Release assets are versioned (<binary>-<tag>-<os>-<arch>); the wildcard
			// matches the latest tag, mirroring scripts/install.sh.
			pattern := fmt.Sprintf("%s-*-%s-%s", config.BinaryName, runtime.GOOS, runtime.GOARCH)
			// Keep the temp file beside the target so the final rename stays on the
			// same filesystem (atomic, and safe to swap a running binary on Unix).
			tempPath := execPath + ".tmp"

			_, err = tui.NewSpinner[any](tui.SpinnerConfig{
				Title: fmt.Sprintf(" Downloading %s...", pattern),
				Type:  tui.SpinnerTypePoints,
			}).Exec(func() (any, error) {
				// No tag argument: gh downloads the latest release.
				ghCmd := shell.NewCommand(
					"gh", "release", "download",
					"--repo", config.Repo,
					"--pattern", pattern,
					"--output", tempPath,
					"--clobber",
				)
				if err := ghCmd.Run(); err != nil {
					return nil, fmt.Errorf("failed to download: %w", err)
				}
				return nil, nil
			})
			if err != nil {
				return err
			}

			if err := os.Chmod(tempPath, 0o755); err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to set executable permissions: %w", err)
			}

			_, err = tui.NewSpinner[any](tui.SpinnerConfig{
				Title: " Verifying downloaded binary...",
				Type:  tui.SpinnerTypePoints,
			}).Exec(func() (any, error) {
				if err := shell.NewCommand(tempPath, "help").Run(); err != nil {
					return nil, err
				}
				return nil, nil
			})
			if err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("downloaded binary verification failed: %w", err)
			}

			if err := os.Rename(tempPath, execPath); err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to replace binary: %w", err)
			}

			// Report success before refreshing completion: completion.Install can exit
			// non-zero when shell completion is not enabled, and that must not mask
			// that the upgrade itself already succeeded.
			tui.StdErrLn("Upgraded to the latest version!")

			if err := completion.Install(cmd); err != nil {
				tui.StdErrF("Warning: failed to install shell completion: %v\n", err)
			} else {
				tui.StdErrLn("Updated shell completion!")
			}

			return nil
		},
	}
}
