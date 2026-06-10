package session

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/spf13/cobra"
)

func SwitchCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch to a tmux session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if target != "" {
				return tmux.SwitchClient(&tmux.SwitchClientParams{TargetSession: target})
			}
			return switchInteractive()
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "session to switch to")

	_ = cmd.RegisterFlagCompletionFunc("target", sessionNameCompletion)

	return cmd
}

// sessionNameCompletion offers visible session names, each annotated with its window count.
func sessionNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	infos, err := listSessionInfos()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	out := make([]string, 0, len(infos))
	for _, info := range infos {
		out = append(out, info.Name+"\t"+sessionSummary(info))
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
