package cmd

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sbchaos/mirage/tui"
)

func NewCmdWindow() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "window",
		Short:   "Inspect window for a job",
		Example: "mirage window",
		Run:     runShowWindow,
	}
	return cmd
}

func runShowWindow(cmd *cobra.Command, args []string) {
	model, err := tui.NewWindow()
	if err != nil {
		fmt.Println(tui.RenderError(fmt.Sprintf("Error starting window command: %s", err)) + "\n")
		return
	}
	if err := tea.NewProgram(model).Start(); err != nil {
		log.Fatal(err)
	}
}
