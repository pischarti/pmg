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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if !stringsContains(out.String(), "skip remove A example.com") {
		t.Fatalf("expected skip remove message, got %q", out.String())
	}
}

func TestSyncDNSRecords_RemoveTXTWithQuotes(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "TXT", Name: "example.com", Content: `"fah-claim=023-02-1e02dfb2-a08d-4bc3-97e8-d7b6db9aa81e"`},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "TXT", Action: ActionRemove, Value: "fah-claim=023-02-1e02dfb2-a08d-4bc3-97e8-d7b6db9aa81e"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.deleteCalls) != 1 || api.deleteCalls[0] != "1" {
		t.Fatalf("expected delete call for record 1, got %#v", api.deleteCalls)
	}
	if !stringsContains(out.String(), "removed TXT example.com") {
		t.Fatalf("expected remove message, got %q", out.String())
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
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

	err := SyncDNSRecords(ctx, api, "example.com", specs, ioDiscard{}, false, false)
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, true, true); err != nil {
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, true, true); err != nil {
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

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if !stringsContains(out.String(), "conflicts with existing record") {
		t.Fatalf("expected conflict message, got %q", out.String())
	}
}

func TestSyncDNSRecords_NilAPI(t *testing.T) {
	ctx := context.Background()
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	err := SyncDNSRecords(ctx, nil, "example.com", specs, ioDiscard{}, false, false)
	if err == nil || !stringsContains(err.Error(), "nil") {
		t.Fatalf("expected nil API error, got %v", err)
	}
}

func TestSyncDNSRecords_NilWriter(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{zoneID: "zone"}
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	err := SyncDNSRecords(ctx, api, "example.com", specs, nil, false, false)
	if err != nil {
		t.Fatalf("expected no error with nil writer, got %v", err)
	}
}

type fakeWriteAPIWithZoneError struct {
	fakeWriteAPI
}

func (f *fakeWriteAPIWithZoneError) ZoneIDByName(string) (string, error) {
	return "", errors.New("zone not found")
}

func TestSyncDNSRecords_ZoneLookupError(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPIWithZoneError{
		fakeWriteAPI: fakeWriteAPI{zoneID: ""},
	}

	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	err := SyncDNSRecords(ctx, api, "example.com", specs, ioDiscard{}, false, false)
	if err == nil || !stringsContains(err.Error(), "zone not found") {
		t.Fatalf("expected zone lookup error, got %v", err)
	}
}

func TestSyncDNSRecords_RemoveWithTable(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4"},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionRemove, Value: "1.2.3.4"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, true); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if !stringsContains(out.String(), "REMOVE") || !stringsContains(out.String(), "example.com") {
		t.Fatalf("expected table output with REMOVE, got %q", out.String())
	}
}

func TestSyncDNSRecords_RemoveDeleteError(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4"},
		},
		deleteErr: errors.New("delete failed"),
	}

	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionRemove, Value: "1.2.3.4"},
	}

	err := SyncDNSRecords(ctx, api, "example.com", specs, ioDiscard{}, false, false)
	if err == nil || !stringsContains(err.Error(), "delete failed") {
		t.Fatalf("expected delete error, got %v", err)
	}
}

func TestSyncDNSRecords_UpdateError(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "5.5.5.5", TTL: 60},
		},
		updateErr: errors.New("update failed"),
	}

	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.1.1.1"},
	}

	err := SyncDNSRecords(ctx, api, "example.com", specs, ioDiscard{}, false, false)
	if err == nil || !stringsContains(err.Error(), "update failed") {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestSyncDNSRecords_CreateError(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID:    "zone",
		createErr: errors.New("create failed"),
	}

	specs := []RecordSpec{
		{Name: "new.example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	err := SyncDNSRecords(ctx, api, "example.com", specs, ioDiscard{}, false, false)
	if err == nil || !stringsContains(err.Error(), "create failed") {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestSyncDNSRecords_RemoveNoValue(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4"},
			{ID: "2", Type: "A", Name: "example.com", Content: "5.6.7.8"},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionRemove},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.deleteCalls) != 2 {
		t.Fatalf("expected 2 delete calls, got %d", len(api.deleteCalls))
	}
}

func TestSyncDNSRecords_SyncMatchingRecordNoUpdate(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 300, Proxied: cf.BoolPtr(true)},
		},
	}

	var out bytes.Buffer
	ttl := 300
	proxied := true
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4", TTL: &ttl, Proxied: &proxied},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.updateCalls) != 0 {
		t.Fatalf("expected no update calls when record matches, got %d", len(api.updateCalls))
	}
}

func TestSyncDNSRecords_EnsureRecordWithTrailingDot(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com.", Content: "1.2.3.4"},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "5.6.7.8"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	// The record should be found and updated (either via syncMatchingRecord or syncExistingRecord)
	// Since content differs, it should update via syncExistingRecord
	if len(api.updateCalls) == 0 && len(api.createCalls) == 0 {
		t.Fatalf("expected update or create call, got updateCalls=%d createCalls=%d", len(api.updateCalls), len(api.createCalls))
	}
}

func TestLoadSyncConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	content := `records: []`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "does not contain any records") {
		t.Fatalf("expected empty records error, got %v", err)
	}
}

func TestLoadSyncConfig_MissingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-name.yaml")
	content := `
records:
  - recordType: A
    action: add
    value: 1.2.3.4
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "name is required") {
		t.Fatalf("expected missing name error, got %v", err)
	}
}

func TestLoadSyncConfig_MissingType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-type.yaml")
	content := `
records:
  - name: example.com
    action: add
    value: 1.2.3.4
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "recordType is required") {
		t.Fatalf("expected missing type error, got %v", err)
	}
}

func TestLoadSyncConfig_MissingAction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-action.yaml")
	content := `
records:
  - name: example.com
    recordType: A
    value: 1.2.3.4
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "action is required") {
		t.Fatalf("expected missing action error, got %v", err)
	}
}

func TestLoadSyncConfig_InvalidAction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid-action.yaml")
	content := `
records:
  - name: example.com
    recordType: A
    action: invalid
    value: 1.2.3.4
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "unsupported action") {
		t.Fatalf("expected invalid action error, got %v", err)
	}
}

func TestLoadSyncConfig_MissingValueForAdd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-value.yaml")
	content := `
records:
  - name: example.com
    recordType: A
    action: add
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "value is required for add actions") {
		t.Fatalf("expected missing value error, got %v", err)
	}
}

func TestLoadSyncConfig_FileNotFound(t *testing.T) {
	_, err := LoadSyncConfig("/nonexistent/file.yaml")
	if err == nil || !stringsContains(err.Error(), "reading sync file") {
		t.Fatalf("expected file read error, got %v", err)
	}
}

func TestLoadSyncConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	content := `invalid: yaml: [`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, err := LoadSyncConfig(path)
	if err == nil || !stringsContains(err.Error(), "parsing sync file") {
		t.Fatalf("expected YAML parse error, got %v", err)
	}
}

func TestSyncDNSRecords_DiffSummaryInOutput(t *testing.T) {
	tests := []struct {
		name         string
		record       cf.DNSRecord
		spec         RecordSpec
		wantContains []string
	}{
		{
			name:         "ttl change only",
			record:       cf.DNSRecord{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 120, Proxied: cf.BoolPtr(false)},
			spec:         RecordSpec{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4", TTL: intPtr(300)},
			wantContains: []string{"ttl=120->300"},
		},
		{
			name:         "proxied change only",
			record:       cf.DNSRecord{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 300, Proxied: cf.BoolPtr(false)},
			spec:         RecordSpec{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4", Proxied: cf.BoolPtr(true)},
			wantContains: []string{"proxied=false->true"},
		},
		{
			name:         "both changes",
			record:       cf.DNSRecord{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 120, Proxied: cf.BoolPtr(false)},
			spec:         RecordSpec{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4", TTL: intPtr(300), Proxied: cf.BoolPtr(true)},
			wantContains: []string{"ttl=120->300", "proxied=false->true"},
		},
		{
			name:         "nil proxied to true",
			record:       cf.DNSRecord{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 300, Proxied: nil},
			spec:         RecordSpec{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4", Proxied: cf.BoolPtr(true)},
			wantContains: []string{"->true"}, // When proxied is nil, it shows as empty current, then ->true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			api := &fakeWriteAPI{
				zoneID:  "zone",
				records: []cf.DNSRecord{tt.record},
			}

			var out bytes.Buffer
			specs := []RecordSpec{tt.spec}

			if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, true, true); err != nil {
				t.Fatalf("SyncDNSRecords returned error: %v", err)
			}

			got := out.String()
			for _, want := range tt.wantContains {
				if !stringsContains(got, want) {
					t.Errorf("output = %q, want contains %q", got, want)
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestSyncDNSRecords_UpdateExistingRecordType(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4", TTL: 60},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "CNAME", Action: ActionAdd, Value: "target.example.com"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, false); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if len(api.updateCalls) != 1 {
		t.Fatalf("expected update call, got %d", len(api.updateCalls))
	}
	if api.updateCalls[0].Type != "CNAME" {
		t.Fatalf("expected type CNAME, got %s", api.updateCalls[0].Type)
	}
	if api.updateCalls[0].Content != "target.example.com" {
		t.Fatalf("expected content target.example.com, got %s", api.updateCalls[0].Content)
	}
}

func TestSyncDNSRecords_AddCnameConflict_DryRun(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID:  "zone",
		records: []cf.DNSRecord{{ID: "cname-id", Type: "CNAME", Name: "example.com", Content: "cname.target"}},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionAdd, Value: "1.2.3.4"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, true, true); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	// In dry-run, we show what would happen - since there's a CNAME, we'd try to update it to A
	// The conflict would only be detected on actual creation, not in dry-run
	if !stringsContains(out.String(), "UPDATE") || !stringsContains(out.String(), "example.com") {
		t.Fatalf("expected UPDATE in dry-run output (would update existing CNAME to A), got %q", out.String())
	}
}

func TestSyncDNSRecords_RemoveWithTableOutput(t *testing.T) {
	ctx := context.Background()
	api := &fakeWriteAPI{
		zoneID: "zone",
		records: []cf.DNSRecord{
			{ID: "1", Type: "A", Name: "example.com", Content: "1.2.3.4"},
		},
	}

	var out bytes.Buffer
	specs := []RecordSpec{
		{Name: "example.com", Type: "A", Action: ActionRemove, Value: "1.2.3.4"},
	}

	if err := SyncDNSRecords(ctx, api, "example.com", specs, &out, false, true); err != nil {
		t.Fatalf("SyncDNSRecords returned error: %v", err)
	}

	if !stringsContains(out.String(), "REMOVE") {
		t.Fatalf("expected REMOVE in table output, got %q", out.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func stringsContains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
