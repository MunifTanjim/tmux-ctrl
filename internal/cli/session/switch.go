package session

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/spf13/cobra"
)

func SwitchCommand() *cobra.Command {
	var (
		target string
		hidden bool
	)

	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch to a tmux session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if target != "" {
				return tmux.SwitchClient(&tmux.SwitchClientParams{TargetSession: target})
			}
			return switchInteractive(hidden)
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "session to switch to")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "include the hidden session")

	_ = cmd.RegisterFlagCompletionFunc("target", sessionNameCompletion)

	return cmd
}

// sessionNameCompletion offers visible session names, each annotated with its window count.
func sessionNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	hidden, _ := cmd.Flags().GetBool("hidden")
	infos, err := listSessionInfos(hidden)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	out := make([]string, 0, len(infos))
	for _, info := range infos {
		out = append(out, info.Name+"\t"+sessionSummary(info))
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
