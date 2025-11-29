package cloudflare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	cf "github.com/cloudflare/cloudflare-go"
	pkgcf "github.com/pischarti/pmg/pkg/cloudflare"
	"github.com/spf13/cobra"
)

type stubAPI struct {
	calls int
}

func (s *stubAPI) ZoneIDByName(string) (string, error) { return "zone", nil }

func (s *stubAPI) ListDNSRecords(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
	_ = ctx
	_ = rc
	_ = params
	s.calls++
	return nil, nil, nil
}

func resetGlobals(t *testing.T) {
	t.Helper()
	clientFactory = pkgcf.NewClientFromEnv
	listRecords = pkgcf.ListDNSRecords
	renderTable = pkgcf.RenderDNSRecordsTable
	loadSyncConfig = pkgcf.LoadSyncConfig
	syncClientFactory = pkgcf.NewWriteClientFromEnv
	syncRecordsFunc = pkgcf.SyncDNSRecords
	syncFilePath = ""
	syncDryRun = false
}

func TestRunList_NoRecords(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	clientFactory = func() (pkgcf.DNSAPI, error) {
		return &stubAPI{}, nil
	}
	listRecords = func(ctx context.Context, api pkgcf.DNSAPI, domain string) ([]cf.DNSRecord, error) {
		return []cf.DNSRecord{}, nil
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := runList(cmd, []string{"example.com"}); err != nil {
		t.Fatalf("runList returned error: %v", err)
	}

	if got := out.String(); got != "no DNS records found for example.com\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunList_RenderRecords(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	clientFactory = func() (pkgcf.DNSAPI, error) {
		return &stubAPI{}, nil
	}
	listRecords = func(ctx context.Context, api pkgcf.DNSAPI, domain string) ([]cf.DNSRecord, error) {
		return []cf.DNSRecord{
			{Type: "A", Name: "example.com", Content: "1.2.3.4"},
		}, nil
	}

	var captured []cf.DNSRecord
	renderTable = func(w io.Writer, records []cf.DNSRecord) {
		captured = append([]cf.DNSRecord{}, records...)
		fmt.Fprintln(w, "rendered")
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := runList(cmd, []string{"example.com"}); err != nil {
		t.Fatalf("runList returned error: %v", err)
	}

	if len(captured) != 1 || captured[0].Name != "example.com" {
		t.Fatalf("expected record captured, got %#v", captured)
	}

	if got := out.String(); got != "rendered\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunList_RequestError6003(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	clientFactory = func() (pkgcf.DNSAPI, error) {
		return &stubAPI{}, nil
	}
	reqErr := cf.NewRequestError(&cf.Error{ErrorCodes: []int{6003}})
	listRecords = func(ctx context.Context, api pkgcf.DNSAPI, domain string) ([]cf.DNSRecord, error) {
		return nil, reqErr
	}

	cmd := &cobra.Command{}

	err := runList(cmd, []string{"example.com"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, reqErr) {
		// Check the message instead to avoid creating multiple request errors.
		if got := err.Error(); !containsAll(got, []string{"6003", "CLOUDFLARE_API_TOKEN", "CLOUDFLARE_API_KEY"}) {
			t.Fatalf("unexpected error message: %s", err)
		}
	}
}

func TestRunList_CFError6003(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	clientFactory = func() (pkgcf.DNSAPI, error) {
		return &stubAPI{}, nil
	}
	apiErr := &cf.Error{ErrorCodes: []int{6003}}
	listRecords = func(ctx context.Context, api pkgcf.DNSAPI, domain string) ([]cf.DNSRecord, error) {
		return nil, fmt.Errorf("wrapped: %w", apiErr)
	}

	cmd := &cobra.Command{}

	err := runList(cmd, []string{"example.com"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got := err.Error(); !containsAll(got, []string{"6003", "invalid request headers", "CLOUDFLARE_API_TOKEN"}) {
		t.Fatalf("unexpected error message: %s", err)
	}
}

func TestRunList_ClientFactoryError(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	clientFactory = func() (pkgcf.DNSAPI, error) {
		return nil, errors.New("boom")
	}

	cmd := &cobra.Command{}
	if err := runList(cmd, []string{"example.com"}); err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

type noopWriteAPI struct{}

func (n *noopWriteAPI) ZoneIDByName(string) (string, error) { return "zone", nil }

func (n *noopWriteAPI) ListDNSRecords(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
	_ = ctx
	_ = rc
	_ = params
	return nil, &cf.ResultInfo{}, nil
}

func (n *noopWriteAPI) CreateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.CreateDNSRecordParams) (cf.DNSRecord, error) {
	_ = ctx
	_ = rc
	_ = params
	return cf.DNSRecord{}, nil
}

func (n *noopWriteAPI) UpdateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error) {
	_ = ctx
	_ = rc
	_ = params
	return cf.DNSRecord{}, nil
}

func (n *noopWriteAPI) DeleteDNSRecord(ctx context.Context, rc *cf.ResourceContainer, recordID string) error {
	_ = ctx
	_ = rc
	_ = recordID
	return nil
}

func TestRunSync_Success(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	syncFilePath = "records.yaml"
	loadSyncConfig = func(path string) ([]pkgcf.RecordSpec, error) {
		if path != "records.yaml" {
			t.Fatalf("expected path records.yaml, got %s", path)
		}
		return []pkgcf.RecordSpec{
			{Name: "example.com", Type: "A", Action: pkgcf.ActionAdd, Value: "1.2.3.4"},
		}, nil
	}
	syncClientFactory = func() (pkgcf.DNSWriteAPI, error) {
		return &noopWriteAPI{}, nil
	}

	syncRecordsFunc = func(ctx context.Context, api pkgcf.DNSWriteAPI, domain string, specs []pkgcf.RecordSpec, w io.Writer, dryRun bool) error {
		if domain != "example.com" {
			t.Fatalf("unexpected domain %s", domain)
		}
		if len(specs) != 1 {
			t.Fatalf("expected one spec, got %d", len(specs))
		}
		if dryRun {
			t.Fatalf("expected dryRun false")
		}
		fmt.Fprintln(w, "synced")
		return nil
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := runSync(cmd, []string{"example.com"}); err != nil {
		t.Fatalf("runSync returned error: %v", err)
	}
	if got := out.String(); got != "synced\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunSync_DecoratedError(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	syncFilePath = "records.yaml"
	loadSyncConfig = func(path string) ([]pkgcf.RecordSpec, error) {
		return []pkgcf.RecordSpec{
			{Name: "example.com", Type: "A", Action: pkgcf.ActionAdd, Value: "1.2.3.4"},
		}, nil
	}
	syncClientFactory = func() (pkgcf.DNSWriteAPI, error) {
		return &noopWriteAPI{}, nil
	}

	reqErr := cf.NewRequestError(&cf.Error{ErrorCodes: []int{6003}})
	syncRecordsFunc = func(ctx context.Context, api pkgcf.DNSWriteAPI, domain string, specs []pkgcf.RecordSpec, w io.Writer, dryRun bool) error {
		if !dryRun {
			t.Fatalf("expected dryRun true")
		}
		return reqErr
	}

	cmd := &cobra.Command{}
	syncDryRun = true
	err := runSync(cmd, []string{"example.com"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); !containsAll(got, []string{"6003", "invalid request headers"}) {
		t.Fatalf("unexpected error message: %s", err)
	}
}

func TestRunSync_LoadError(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	syncFilePath = "records.yaml"
	loadSyncConfig = func(path string) ([]pkgcf.RecordSpec, error) {
		return nil, errors.New("boom")
	}

	cmd := &cobra.Command{}
	if err := runSync(cmd, []string{"example.com"}); err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestRunSync_DryRunFlag(t *testing.T) {
	t.Cleanup(func() { resetGlobals(t) })

	syncFilePath = "records.yaml"
	loadSyncConfig = func(path string) ([]pkgcf.RecordSpec, error) {
		return []pkgcf.RecordSpec{
			{Name: "example.com", Type: "A", Action: pkgcf.ActionAdd, Value: "1.2.3.4"},
		}, nil
	}
	syncClientFactory = func() (pkgcf.DNSWriteAPI, error) {
		return &noopWriteAPI{}, nil
	}

	calledDryRun := false
	syncRecordsFunc = func(ctx context.Context, api pkgcf.DNSWriteAPI, domain string, specs []pkgcf.RecordSpec, w io.Writer, dryRun bool) error {
		if !dryRun {
			t.Fatalf("expected dryRun true")
		}
		calledDryRun = true
		return nil
	}

	cmd := &cobra.Command{}
	syncDryRun = true
	if err := runSync(cmd, []string{"example.com"}); err != nil {
		t.Fatalf("runSync returned error: %v", err)
	}
	if !calledDryRun {
		t.Fatalf("expected syncRecordsFunc to be called with dryRun")
	}
}

func containsAll(s string, substrings []string) bool {
	for _, sub := range substrings {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
