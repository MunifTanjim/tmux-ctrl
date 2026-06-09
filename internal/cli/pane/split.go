package pane

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func SplitCommand() *cobra.Command {
	var (
		paneID    string
		direction string
		size      string
		edge      bool
		env       []string
		template  string
		noFocus   bool
	)

	cmd := &cobra.Command{
		Use:   "split [command [args...]]",
		Short: "Split a pane to create a new one",
		Long: "Split a pane to create a new one.\n\n" +
			"Any trailing arguments are run as a command in the new pane instead of\n" +
			"the default shell. Flags must precede the command; everything after the\n" +
			"first non-flag argument (or after --) is treated as the command.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			if err := ensureWindowUnzoomed(paneID); err != nil {
				return err
			}

			o := splitOptions{
				paneID:    paneID,
				direction: direction,
				size:      size,
				template:  template,
				command:   args,
				env:       env,
				focus:     !noFocus,
			}

			// Corners are inherently an edge placement; assume --edge for them.
			var out string
			if edge || isCornerDirection(direction) {
				out, err = splitPaneToEdge(o)
			} else {
				out, err = splitPane(o)
			}
			if err != nil {
				return err
			}

			if out != "" {
				tui.StdOutLn(out)
			}
			return nil
		},
	}

	// Stop flag parsing at the first positional so the command's own flags
	// (e.g. `pane split nvim -u NONE`) are passed through rather than consumed.
	cmd.Flags().SetInterspersed(false)

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "pane to split (default: current pane)")
	cmd.Flags().StringVarP(&direction, "direction", "d", "bottom", "split direction: "+directionsHint)
	cmd.Flags().StringVar(&size, "size", "", "size of the new pane (lines, columns, percentage, or 0<fraction<1 as percentage)")
	cmd.Flags().BoolVar(&edge, "edge", false, "split at the window edge in --direction, spanning the full side")
	cmd.Flags().StringArrayVar(&env, "env", nil, "set an environment variable (VAR=value) in the new pane; repeatable")
	cmd.Flags().StringVar(&template, "template", "", "tmux format string to print info about the new pane")
	cmd.Flags().BoolVar(&noFocus, "no-focus", false, "do not focus the new pane")

	cmd.MarkFlagsMutuallyExclusive("edge", "pane-id")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(false))
	_ = cmd.RegisterFlagCompletionFunc("direction", directionCompletion)
	_ = cmd.RegisterFlagCompletionFunc("size", noFileCompletion)
	_ = cmd.RegisterFlagCompletionFunc("env", noFileCompletion)
	_ = cmd.RegisterFlagCompletionFunc("template", noFileCompletion)

	return cmd
}
