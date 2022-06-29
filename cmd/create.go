package cmd

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sbchaos/mirage/tui"
)

func NewCmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new job for optimus",
		Example: "mirage create",
		Run:     runCreate,
	}
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) {
	model, err := tui.NewCreateModel()
	if err != nil {
		fmt.Println(tui.RenderError(fmt.Sprintf("Error starting create command: %s", err)) + "\n")
		return
	}
	if err := tea.NewProgram(model).Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(tui.BoldStyle.Copy().Foreground(tui.Green).Render(fmt.Sprintf("🎉 Done!  Your job has been created in ./%s", "dir")))
}
