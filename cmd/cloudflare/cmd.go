package cloudflare

import (
	"context"
	"fmt"

	"github.com/pischarti/pmg/pkg/cloudflare"
	"github.com/pischarti/pmg/pkg/utils"
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

var syncCmd = &cobra.Command{
	Use:   "sync <domain>",
	Short: "Sync DNS records from a YAML specification",
	Args:  cobra.ExactArgs(1),
	RunE:  runSync,
}

var (
	clientFactory     = cloudflare.NewClientFromEnv
	listRecords       = cloudflare.ListDNSRecords
	renderTable       = cloudflare.RenderDNSRecordsTable
	loadSyncConfig    = cloudflare.LoadSyncConfig
	syncClientFactory = cloudflare.NewWriteClientFromEnv
	syncRecordsFunc   = cloudflare.SyncDNSRecords
)

var (
	syncFilePath  string
	syncDryRun    bool
	syncShowTable bool
)

func init() {
	Cmd.AddCommand(listCmd)
	syncCmd.Flags().StringVarP(&syncFilePath, "file", "f", "", "Path to YAML file describing DNS records")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without modifying Cloudflare")
	syncCmd.Flags().BoolVar(&syncShowTable, "table", true, "Render a tabular summary of sync actions")
	cobra.CheckErr(syncCmd.MarkFlagRequired("file"))
	Cmd.AddCommand(syncCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	domain := args[0]

	api, err := clientFactory()
	if err != nil {
		return err
	}

	ctx := context.Background()

	records, err := listRecords(ctx, api, domain)
	if err != nil {
		return utils.DecorateAuthError(err)
	}

	if len(records) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no DNS records found for %s\n", domain)
		return nil
	}
	renderTable(cmd.OutOrStdout(), records)
	return nil
}

func runSync(cmd *cobra.Command, args []string) error {
	domain := args[0]

	specs, err := loadSyncConfig(syncFilePath)
	if err != nil {
		return err
	}

	api, err := syncClientFactory()
	if err != nil {
		return err
	}

	ctx := context.Background()
	showTable := syncShowTable || syncDryRun
	if err := syncRecordsFunc(ctx, api, domain, specs, cmd.OutOrStdout(), syncDryRun, showTable); err != nil {
		return utils.DecorateAuthError(err)
	}

	return nil
}
