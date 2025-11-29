package cmd

import (
	"github.com/pischarti/pmg/cmd/cloudflare"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "pmg",
	Short: "pmg is a tool for managing pmg",
	Long:  `pmg is a tool for managing pmg`,
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.AddCommand(cloudflare.Cmd)
}
