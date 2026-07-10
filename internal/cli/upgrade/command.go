package upgrade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/MunifTanjim/tmux-ctrl/internal/cli/completion"
	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
	"github.com/spf13/cobra"
)

// assetName builds the release-asset filename (<binary>-<tag>-<os>-<arch>).
func assetName(tag, goos, goarch string) string {
	return fmt.Sprintf("%s-%s-%s-%s", config.BinaryName, tag, goos, goarch)
}

// latestTag resolves the latest release tag, preferring gh when available and
// falling back to the public GitHub API otherwise.
func latestTag(ctx context.Context, useGH bool) (string, error) {
	if useGH {
		cmd := shell.NewCommand("gh", "release", "view", "--repo", config.Repo, "--json", "tagName", "--jq", ".tagName")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("%w: %s", err, cmd.StdErr().TrimSpace())
		}
		tag := cmd.StdOut().TrimSpace().String()
		if tag == "" {
			return "", errors.New("release tag missing in response")
		}
		return tag, nil
	}
	return latestReleaseTag(ctx, &http.Client{Timeout: 5 * time.Minute})
}

func latestReleaseTag(ctx context.Context, client *http.Client) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", config.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", errors.New("release tag missing in response")
	}
	return release.TagName, nil
}

// downloadAsset writes the named release asset to dst, preferring gh when
// available and falling back to a direct download otherwise.
func downloadAsset(ctx context.Context, useGH bool, tag, name, dst string) error {
	if useGH {
		dl := shell.NewCommand(
			"gh", "release", "download", tag,
			"--repo", config.Repo,
			"--pattern", name,
			"--output", dst,
			"--clobber",
		)
		if err := dl.Run(); err != nil {
			return fmt.Errorf("%w: %s", err, dl.StdErr().TrimSpace())
		}
		return os.Chmod(dst, 0o755)
	}
	return downloadReleaseAsset(ctx, &http.Client{Timeout: 5 * time.Minute}, tag, name, dst)
}

func downloadReleaseAsset(ctx context.Context, client *http.Client, tag, name, dst string) error {
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", config.Repo, tag, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			useGH := util.HasTool("gh")

			tag, err := latestTag(ctx, useGH)
			if err != nil {
				return fmt.Errorf("failed to resolve latest release: %w", err)
			}

			name := assetName(tag, runtime.GOOS, runtime.GOARCH)
			// Keep the temp file beside the target so the final rename stays on the
			// same filesystem (atomic, and safe to swap a running binary on Unix).
			tempPath := execPath + ".tmp"

			_, err = tui.NewSpinner[any](tui.SpinnerConfig{
				Title: fmt.Sprintf(" Downloading %s...", name),
				Type:  tui.SpinnerTypePoints,
			}).Exec(func() (any, error) {
				if err := downloadAsset(ctx, useGH, tag, name, tempPath); err != nil {
					return nil, fmt.Errorf("failed to download: %w", err)
				}
				return nil, nil
			})
			if err != nil {
				os.Remove(tempPath)
				return err
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
