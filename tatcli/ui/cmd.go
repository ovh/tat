package ui

import (
	"github.com/gizak/termui"
	"github.com/ovh/tat/tatcli/internal"
	"github.com/spf13/cobra"
)

// Cmd dashboard
var Cmd = &cobra.Command{
	Use:   "ui",
	Short: "Interactive mode: tatcli ui [<topic>] [<command>,<command>,...]",
	Long: `Interactive mode: tatcli ui [<topic>] [<command>,<command>,...]
Example:

	tatcli ui /YourTopic/SubTopic /run AA,BB /hide-usernames /hide-top
	tatcli ui /YourTopic/SubTopic /split label:open label:doing label:done /mode run /save
	tatcli ui /YourTopic/SubTopic /run AA,BB /hide-usernames /hide-bottom /save

	`,
	Aliases: []string{"d", "dashboard"},
	Run: func(cmd *cobra.Command, args []string) {
		runUI(args)
	},
}

func runUI(args []string) {
	ui := &tatui{}

	// This forces a client init before starting the UI
	// This way, if the client needs to ask for credentials, it will be done now
	internal.Client().Version()

	ui.init(args)
	ui.draw(0)

	defer termui.Close()
	termui.Loop()
}
