package cloudflare

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cf "github.com/cloudflare/cloudflare-go"
)

type fakeWriteAPI struct {
	zoneID      string
	records     []cf.DNSRecord
	createCalls []cf.CreateDNSRecordParams
	deleteCalls []string
	updateCalls []cf.UpdateDNSRecordParams
	listErr     error
	createErr   error
	deleteErr   error
	updateErr   error
}

func (f *fakeWriteAPI) ZoneIDByName(string) (string, error) { return f.zoneID, nil }

func (f *fakeWriteAPI) ListDNSRecords(ctx context.Context, rc *cf.ResourceContainer, params cf.ListDNSRecordsParams) ([]cf.DNSRecord, *cf.ResultInfo, error) {
	_ = ctx
	_ = rc
	_ = params
	if f.listErr != nil {
		return nil, nil, f.listErr
	}
	var records []cf.DNSRecord
	for _, record := range f.records {
		if params.Type != "" && !strings.EqualFold(params.Type, record.Type) {
			continue
		}
		if params.Name != "" && !strings.EqualFold(params.Name, record.Name) {
			continue
		}
		records = append(records, record)
	}
	return records, &cf.ResultInfo{}, nil
}

func (f *fakeWriteAPI) CreateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.CreateDNSRecordParams) (cf.DNSRecord, error) {
	_ = ctx
	_ = rc
	if f.createErr != nil {
		return cf.DNSRecord{}, f.createErr
	}
	f.createCalls = append(f.createCalls, params)
	record := cf.DNSRecord{
		ID:      params.Name + "|" + params.Type,
		Name:    params.Name,
		Type:    params.Type,
		Content: params.Content,
		TTL:     params.TTL,
	}
	if params.Proxied != nil {
		record.Proxied = params.Proxied
	}
	f.records = append(f.records, record)
	return record, nil
}

func (f *fakeWriteAPI) DeleteDNSRecord(ctx context.Context, rc *cf.ResourceContainer, recordID string) error {
	_ = ctx
	_ = rc
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleteCalls = append(f.deleteCalls, recordID)
	filtered := f.records[:0]
	for _, record := range f.records {
		if record.ID != recordID {
			filtered = append(filtered, record)
		}
	}
	f.records = filtered
	return nil
}

func (f *fakeWriteAPI) UpdateDNSRecord(ctx context.Context, rc *cf.ResourceContainer, params cf.UpdateDNSRecordParams) (cf.DNSRecord, error) {
	_ = ctx
	_ = rc
	if f.updateErr != nil {
		return cf.DNSRecord{}, f.updateErr
	}
	f.updateCalls = append(f.updateCalls, params)
	for i := range f.records {
		if f.records[i].ID == params.ID {
			f.records[i].Content = params.Content
			f.records[i].TTL = params.TTL
			if params.Proxied != nil {
				f.records[i].Proxied = params.Proxied
			}
			break
		}
	}
	return cf.DNSRecord{
		ID:      params.ID,
		Type:    params.Type,
		Name:    params.Name,
		Content: params.Content,
		TTL:     params.TTL,
	}, nil
}

func TestLoadSyncConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "records.yaml")
	content := `
records:
  - name: example.com
    recordType: A
    action: add
    value: 1.2.3.4
    ttl: 300
    proxied: true
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	records, err := LoadSyncConfig(path)
	if err != nil {
		t.Fatalf("LoadSyncConfig returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0].Type != "A" || records[0].Action != ActionAdd {
		t.Fatalf("unexpected record parsed: %#v", records[0])
	}
	if records[0].TTL == nil || *records[0].TTL != 300 {
		t.Fatalf("expected ttl 300, got %#v", records[0].TTL)
	}
	if records[0].Proxied == nil || !*records[0].Proxied {
		t.Fatalf("expected proxied true, got %#v", records[0].Proxied)
	}
}

func TestSyncDNSRecords_AddAndSkip(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "existing", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 1},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
		{Name: "new.example.com", Type: "A", Action: ActionAdd, Value: "5.6.7.8"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.createCalls) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(api.createCalls))
	}
	if api.createCalls[0].Name != "new.example.com" {
		t.Fatalf("unexpected create params: %#v", api.createCalls[0])
	}

	if !stringsContains(out.String(), "skip add A example.com") {
		t.Fatalf("expected skip message in output: %q", out.String())
	}
	if !stringsContains(out.String(), "added A new.example.com") {
		t.Fatalf("expected add message in output: %q", out.String())
	}
}

func TestSyncDNSRecords_Remove(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "TXT", Name: "example.com", Content: "foo"},
			{ID: "2", Type: "TXT", Name: "example.com", Content: "bar"},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "TXT", Action: ActionRemove, Value: "bar"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.deleteCalls) != 1 || api.deleteCalls[0] != "2" {
		t.Fatalf("expected delete call for record 2, got %#v", api.deleteCalls)
	}
	if !stringsContains(out.String(), "removed TXT example.com -> bar") {
		t.Fatalf("expected remove message, got %q", out.String())
	}
}

func TestSyncDNSRecords_RemoveNoMatch(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID:  "zone",
		records: []cf.DNSRecord{},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionRemove},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if !stringsContains(out.String(), "skip remove A example.com") {
		t.Fatalf("expected skip remove message, got %q", out.String())
	}
}

func TestSyncDNSRecords_UpdateMatchingRecord(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 120},
		},
	}

	var out bytes.Buffer
	ttl := 300
	proxied := true
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4", TTL: &ttl, Proxied: &proxied},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.updateCalls) != 1 {
		t.Fatalf("expected update call, got %d", len(api.updateCalls))
	}
	upd := api.updateCalls[0]
	if upd.TTL != ttl {
		t.Fatalf("expected ttl %d, got %d", ttl, upd.TTL)
	}
	if upd.Proxied == nil || !*upd.Proxied {
		t.Fatalf("expected proxied true, got %#v", upd.Proxied)
	}
	if !stringsContains(out.String(), "updated A example.com -> 1.2.3.4") {
		t.Fatalf("expected update message, got %q", out.String())
	}
}

func TestSyncDNSRecords_UpdateExistingRecordContent(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "5.5.5.5", TTL: 60},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.1.1.1"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.updateCalls) != 1 {
		t.Fatalf("expected update call, got %d", len(api.updateCalls))
	}
	if api.updateCalls[0].Content != "1.1.1.1" {
		t.Fatalf("expected content update to 1.1.1.1, got %s", api.updateCalls[0].Content)
	}
}

func TestSyncDNSRecords_ListError(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID:  "zone",
		listErr: errors.New("boom"),
	}

	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	err := SyncDNSRecords(ctx, api, "example.com", specs, ioDiscard{}, false)
	if err == nil || !stringsContains(err.Error(), "boom") {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestSyncDNSRecords_DryRunAddRemove(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 1},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
		{Name: "new.example.com", Type: "A", Action: ActionAdd, Value: "5.6.7.8"},
		{Name: "example.com", Type: "A", Action: ActionRemove, Value: "1.2.3.4"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, true); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.createCalls) != 0 {
		t.Fatalf("expected no create calls in dry run, got %d", len(api.createCalls))
	}
	if len(api.deleteCalls) != 0 {
		t.Fatalf("expected no delete calls in dry run, got %d", len(api.deleteCalls))
	}

	outStr := out.String()
	if !stringsContains(outStr, "SKIP") || !stringsContains(outStr, "example.com") {
		t.Fatalf("expected dry-run skip row, got %q", outStr)
	}
	if !stringsContains(outStr, "ADD") || !stringsContains(outStr, "new.example.com") {
		t.Fatalf("expected dry-run add row, got %q", outStr)
	}
	if !stringsContains(outStr, "REMOVE") || !stringsContains(outStr, "1.2.3.4") {
		t.Fatalf("expected dry-run remove row, got %q", outStr)
	}
}

func TestSyncDNSRecords_DryRunUpdateTable(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "5.5.5.5", TTL: 120},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.1.1.1"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, true); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.updateCalls) != 0 {
		t.Fatalf("expected no updates in dry run, got %d", len(api.updateCalls))
	}

	outStr := out.String()
	if !stringsContains(outStr, "UPDATE") || !stringsContains(outStr, "1.1.1.1") {
		t.Fatalf("expected update row in dry-run output, got %q", outStr)
	}
}

func TestSyncDNSRecords_AddCnameConflict(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID:    "zone",
		createErr: cf.NewRequestError(&cf.Error{ErrorCodes: []int{81054}}),
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if !stringsContains(out.String(), "conflicts with existing record") {
		t.Fatalf("expected conflict message, got %q", out.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func stringsContains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
