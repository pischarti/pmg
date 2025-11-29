package cloudflare

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	cf "github.com/cloudflare/cloudflare-go"
)

type mockDNSAPI struct {
	zoneID     string
	zoneErr    error
	records    []cf.DNSRecord
	listErr    error
	calledWith struct {
		domain string
	}
}

func (m *mockDNSAPI) ZoneIDByName(domain string) (string, error) {
	m.calledWith.domain = domain
	if m.zoneErr != nil {
		return "", m.zoneErr
	}
	return m.zoneID, nil
}

func (m *mockDNSAPI) ListDNSRecords(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return append([]cf.DNSRecord(nil), m.records...), &cf.ResultInfo{}, nil
}

func TestListDNSRecords_SortsResults(t *testing.T) {
	mock := &mockDNSAPI{
		zoneID: "zone-id",
		records: []cf.DNSRecord{
			{Type: "TXT", Name: "b.example.com"},
			{Type: "A", Name: "a.example.com"},
			{Type: "AAAA", Name: "a.example.com"},
		},
	}

	got, err := ListDNSRecords(context.Background(), mock, "example.com")
	if err != nil {
		t.Fatalf("ListDNSRecords returned error: %v", err)
	}

	if mock.calledWith.domain != "example.com" {
		t.Fatalf("expected domain %q, got %q", "example.com", mock.calledWith.domain)
	}

	want := []cf.DNSRecord{
		{Type: "A", Name: "a.example.com"},
		{Type: "AAAA", Name: "a.example.com"},
		{Type: "TXT", Name: "b.example.com"},
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d records, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i].Type != want[i].Type || got[i].Name != want[i].Name {
			t.Fatalf("record %d mismatch: got %+v want %+v", i, got[i], want[i])
		}
	}
}

func TestListDNSRecords_ZoneError(t *testing.T) {
	mock := &mockDNSAPI{
		zoneErr: errors.New("boom"),
	}

	_, err := ListDNSRecords(context.Background(), mock, "example.com")
	if err == nil || !strings.Contains(err.Error(), "example.com") {
		t.Fatalf("expected error mentioning domain, got %v", err)
	}
}

func TestListDNSRecords_ListError(t *testing.T) {
	mock := &mockDNSAPI{
		zoneID:  "zone-id",
		listErr: errors.New("failed list"),
	}

	_, err := ListDNSRecords(context.Background(), mock, "example.com")
	if err == nil || !strings.Contains(err.Error(), "example.com") {
		t.Fatalf("expected error mentioning domain, got %v", err)
	}
}

func TestListDNSRecords_NilAPI(t *testing.T) {
	_, err := ListDNSRecords(context.Background(), nil, "example.com")
	if err == nil || !strings.Contains(err.Error(), "cloudflare API client is nil") {
		t.Fatalf("expected nil client error, got %v", err)
	}
}

func TestRenderDNSRecordsTable(t *testing.T) {
	var buf bytes.Buffer
	proxied := true
	records := []cf.DNSRecord{
		{
			Type:    "A",
			Name:    "a.example.com",
			Content: "1.2.3.4",
			TTL:     1,
			Proxied: &proxied,
		},
	}

	RenderDNSRecordsTable(&buf, records)

	out := buf.String()
	for _, expected := range []string{"TYPE", "A", "a.example.com", "1.2.3.4", "auto", "true"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected output to contain %q, got %q", expected, out)
		}
	}
}
