package cloudflare

import (
	"context"
	"fmt"
	"os"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
)

// DNSWriteAPI extends DNSAPI with mutating operations required for record syncing.
type DNSWriteAPI interface {
	DNSAPI
	CreateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.CreateDNSRecordParams) (cf.DNSRecord, error)
	UpdateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, rc *cf.ResourceContainer, recordID string) error
}

// NewClientFromEnv constructs a DNSAPI using standard Cloudflare environment variables.
// Supports authentication via API token or API key/email pair.
func NewClientFromEnv() (DNSAPI, error) {
	return newRawClientFromEnv()
}

// NewWriteClientFromEnv constructs a DNSWriteAPI capable of mutating DNS resources.
func NewWriteClientFromEnv() (DNSWriteAPI, error) {
	return newRawClientFromEnv()
}

func newRawClientFromEnv() (*cf.API, error) {
	token := strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN"))
	apiKey := strings.TrimSpace(os.Getenv("CLOUDFLARE_API_KEY"))
	apiEmail := strings.TrimSpace(os.Getenv("CLOUDFLARE_API_EMAIL"))

	switch {
	case token != "":
		api, err := cf.NewWithAPIToken(token)
		if err != nil {
			return nil, fmt.Errorf("creating Cloudflare client with API token: %w", err)
		}
		return api, nil
	case apiKey != "" && apiEmail != "":
		api, err := cf.New(apiKey, apiEmail)
		if err != nil {
			return nil, fmt.Errorf("creating Cloudflare client with API key/email: %w", err)
		}
		return api, nil
	case apiKey != "" && apiEmail == "":
		return nil, fmt.Errorf("CLOUDFLARE_API_EMAIL environment variable must be set when using CLOUDFLARE_API_KEY")
	case apiEmail != "" && apiKey == "":
		return nil, fmt.Errorf("CLOUDFLARE_API_KEY environment variable must be set when using CLOUDFLARE_API_EMAIL")
	default:
		return nil, fmt.Errorf("set CLOUDFLARE_API_TOKEN or both CLOUDFLARE_API_KEY and CLOUDFLARE_API_EMAIL to authenticate with Cloudflare")
	}
}
