package cloudflare

import (
	"context"
	"errors"
	"fmt"

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

var (
	clientFactory = cloudflare.NewClientFromEnv
	listRecords   = cloudflare.ListDNSRecords
	renderTable   = cloudflare.RenderDNSRecordsTable
)

func init() {
	Cmd.AddCommand(listCmd)
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
		return decorateAuthError(err)
	}

	if len(records) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no DNS records found for %s\n", domain)
		return nil
	}
	renderTable(cmd.OutOrStdout(), records)
	return nil
}

func decorateAuthError(err error) error {
	var reqErr cf.RequestError
	if errors.As(err, &reqErr) && reqErr.InternalErrorCodeIs(6003) {
		return fmt.Errorf("%w (cloudflare error 6003: invalid request headers); ensure you are using either a valid API token (CLOUDFLARE_API_TOKEN) or an API key plus email (CLOUDFLARE_API_KEY and CLOUDFLARE_API_EMAIL)", err)
	}

	var apiErr *cf.Error
	if errors.As(err, &apiErr) && apiErr.InternalErrorCodeIs(6003) {
		return fmt.Errorf("%w (cloudflare error 6003: invalid request headers); ensure you are using either a valid API token (CLOUDFLARE_API_TOKEN) or an API key plus email (CLOUDFLARE_API_KEY and CLOUDFLARE_API_EMAIL)", err)
	}

	return err
}
