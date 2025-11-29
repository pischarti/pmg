package cloudflare

import (
	"context"
	"fmt"
	"io"
	"sort"

	cf "github.com/cloudflare/cloudflare-go"
	prettytbl "github.com/jedib0t/go-pretty/v6/table"
)

// DNSAPI captures the subset of Cloudflare client capabilities needed for DNS listings.
type DNSAPI interface {
	ZoneIDByName(string) (string, error)
	ListDNSRecords(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error)
}

// ListDNSRecords fetches all DNS records for the provided domain and returns them sorted by name and type.
func ListDNSRecords(ctx context.Context, api DNSAPI, domain string) ([]cf.DNSRecord, error) {
	if api == nil {
		return nil, fmt.Errorf("cloudflare API client is nil")
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return nil, fmt.Errorf("locating zone %q: %w", domain, err)
	}

	rc := cf.ZoneIdentifier(zoneID)

	records, _, err := api.ListDNSRecords(ctx, rc, cf.ListDNSRecordsParams{})
	if err != nil {
		return nil, fmt.Errorf("fetching DNS records for %q: %w", domain, err)
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].Name == records[j].Name {
			return records[i].Type < records[j].Type
		}
		return records[i].Name < records[j].Name
	})

	return records, nil
}

// RenderDNSRecordsTable writes a formatted table of DNS records to the provided writer.
func RenderDNSRecordsTable(w io.Writer, records []cf.DNSRecord) {
	tw := prettytbl.NewWriter()
	tw.SetOutputMirror(w)
	tw.AppendHeader(prettytbl.Row{"TYPE", "NAME", "CONTENT", "TTL", "PROXIED"})

	for _, record := range records {
		ttl := fmt.Sprintf("%d", record.TTL)
		if record.TTL == 1 {
			ttl = "auto"
		}

		proxied := "false"
		if record.Proxied != nil && *record.Proxied {
			proxied = "true"
		}

		tw.AppendRow(prettytbl.Row{
			record.Type,
			record.Name,
			record.Content,
			ttl,
			proxied,
		})
	}

	tw.Render()
}
