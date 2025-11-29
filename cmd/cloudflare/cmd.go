package cloudflare

import (
	"context"
	"fmt"
	"os"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/pischarti/pmg/pkg/cloudflare"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cloudflare",
	Short: "Manage Cloudflare resources",
	Long:  `Commands for inspecting and managing Cloudflare resources.`,
}

var listCmd = &cobra.Command{
	Use:   "list <domain>",
	Short: "List DNS records for a domain",
	Args:  cobra.ExactArgs(1),
	RunE:  runList,
}

func init() {
	Cmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	domain := args[0]

	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	if token == "" {
		return fmt.Errorf("CLOUDFLARE_API_TOKEN environment variable must be set")
	}

	api, err := cf.NewWithAPIToken(token)
	if err != nil {
		return fmt.Errorf("creating Cloudflare client: %w", err)
	}

	ctx := context.Background()

	records, err := cloudflare.ListDNSRecords(ctx, api, domain)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no DNS records found for %s\n", domain)
		return nil
	}
	cloudflare.RenderDNSRecordsTable(cmd.OutOrStdout(), records)
	return nil
}
