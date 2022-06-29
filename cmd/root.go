package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const longDescription = `
    __  ___  _                                
   /  |/  / (_) _____  ____ _  ____ _  ___ 
  / /|_/ / / / / ___/ / __  / / __  / / _ \
 / /  / / / / / /    / /_/ / / /_/ / /  __/
/_/  /_/ /_/ /_/     \__,_/  \__, /  \___/ 
                            /____/         
`

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "mirage",
		Short: "An alternative cli to interact with optimus",
		Long:  longDescription,
	}

	// Register Top Level Commands
	rootCmd.AddCommand(NewCmdCreate())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
