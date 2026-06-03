package cmd

import (
	"strings"
	"time"

	"github.com/MunifTanjim/tmux-ctrl/internal/cache"
	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/MunifTanjim/tmux-ctrl/internal/version"
	"github.com/spf13/cobra"
)

type latestVersionInfo struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
}

var versionCache = cache.New(&cache.Config[latestVersionInfo]{
	Format: "json",
	Prefix: "version",
	TTL:    2 * time.Hour,
})

func getLatestVersion() *latestVersionInfo {
	cached, _ := versionCache.Get("latest")
	if cached != nil {
		return cached
	}

	info, err := tui.NewSpinner[*latestVersionInfo](tui.SpinnerConfig{
		Title: " Checking for updates...",
		Type:  tui.SpinnerTypePoints,
	}).Exec(func() (*latestVersionInfo, error) {
		releaseCmd := shell.NewCommand("gh", "api", "repos/"+config.Repo+"/releases/latest", "--jq", "{tag_name,published_at}")
		if err := releaseCmd.Run(); err != nil {
			return nil, err
		}
		info := latestVersionInfo{}
		if err := releaseCmd.StdOut().TrimSpace().JSONUnmarshal(&info); err != nil {
			return nil, err
		}
		versionCache.Set("latest", info)
		return &info, nil
	})
	if err != nil {
		tui.StdErrF("failed to check for updates: %v\n", err)
	}
	return info
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			tui.StdOutF("tmux-ctrl version %s\n", version.Version)
			tui.StdOutLn("https://github.com/" + config.Repo + "/releases/" + version.Version)

			currentVersion := strings.TrimSuffix(version.Version, "-dirty")
			if latest := getLatestVersion(); latest != nil && currentVersion != latest.TagName {
				tui.StdErrLn()
				tui.StdErrF("A newer version is available! (%s, released %s)\n", latest.TagName, latest.PublishedAt.Format(time.DateTime))
				tui.StdErrLn("Visit the releases page to update: https://github.com/" + config.Repo + "/releases/latest")
			}
		},
	})
}
