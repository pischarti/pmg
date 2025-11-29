package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	table "github.com/jedib0t/go-pretty/v6/table"
	"github.com/pischarti/pmg/pkg/utils"
	"gopkg.in/yaml.v3"
)

// RecordAction represents a supported sync action.
type RecordAction string

const (
	// ActionAdd ensures the record exists (creating if necessary).
	ActionAdd RecordAction = "add"
	// ActionRemove removes matching records.
	ActionRemove RecordAction = "remove"
)

// RecordSpec captures a desired DNS record state from a sync file.
type RecordSpec struct {
	Name    string       `yaml:"name"`
	Type    string       `yaml:"recordType"`
	Action  RecordAction `yaml:"action"`
	Value   string       `yaml:"value"`
	TTL     *int         `yaml:"ttl,omitempty"`
	Proxied *bool        `yaml:"proxied,omitempty"`
}

type syncFile struct {
	Records []RecordSpec `yaml:"records"`
}

// LoadSyncConfig loads and validates record specifications from a YAML file.
func LoadSyncConfig(path string) ([]RecordSpec, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading sync file %q: %w", path, err)
	}

	var cfg syncFile
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parsing sync file %q: %w", path, err)
	}

	if len(cfg.Records) == 0 {
		return nil, fmt.Errorf("sync file %q does not contain any records", path)
	}

	for i := range cfg.Records {
		cfg.Records[i].Action = RecordAction(strings.ToLower(string(cfg.Records[i].Action)))
		cfg.Records[i].Type = strings.ToUpper(cfg.Records[i].Type)

		if cfg.Records[i].Name == "" {
			return nil, fmt.Errorf("record %d: name is required", i)
		}
		if cfg.Records[i].Type == "" {
			return nil, fmt.Errorf("record %d (%s): recordType is required", i, cfg.Records[i].Name)
		}
		if cfg.Records[i].Action == "" {
			return nil, fmt.Errorf("record %d (%s): action is required", i, cfg.Records[i].Name)
		}
		switch cfg.Records[i].Action {
		case ActionAdd, ActionRemove:
			// ok
		default:
			return nil, fmt.Errorf("record %d (%s): unsupported action %q", i, cfg.Records[i].Name, cfg.Records[i].Action)
		}
		if cfg.Records[i].Value == "" && cfg.Records[i].Action == ActionAdd {
			return nil, fmt.Errorf("record %d (%s): value is required for add actions", i, cfg.Records[i].Name)
		}
	}

	return cfg.Records, nil
}

// SyncDNSRecords applies the desired record specifications to Cloudflare.
func SyncDNSRecords(ctx context.Context, api DNSWriteAPI, domain string, specs []RecordSpec, out io.Writer, dryRun bool, showTable bool) error {
	if api == nil {
		return errors.New("cloudflare API client is nil")
	}
	if out == nil {
		out = io.Discard
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return fmt.Errorf("locating zone %q: %w", domain, err)
	}
	rc := cf.ZoneIdentifier(zoneID)

	var dryTable table.Writer
	if dryRun || showTable {
		dryTable = table.NewWriter()
		dryTable.SetOutputMirror(out)
		dryTable.AppendHeader(table.Row{"ACTION", "TYPE", "NAME", "CURRENT", "DESIRED", "DETAIL"})
	}

	for _, spec := range specs {
		switch spec.Action {
		case ActionAdd:
			if err := ensureRecord(ctx, api, rc, spec, out, dryRun, dryTable); err != nil {
				return err
			}
		case ActionRemove:
			if err := removeRecord(ctx, api, rc, spec, out, dryRun, dryTable); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported action %q for record %s", spec.Action, spec.Name)
		}
	}

	if dryTable != nil {
		dryTable.Render()
	}

	return nil
}

func ensureRecord(ctx context.Context, api DNSWriteAPI, rc *cf.ResourceContainer, spec RecordSpec, out io.Writer, dryRun bool, dryTable table.Writer) error {
	specName := trimTrailingDot(spec.Name)
	desiredValue := normalizeRecordValue(spec.Value)

	records, _, err := api.ListDNSRecords(ctx, rc, cf.ListDNSRecordsParams{
		Type: spec.Type,
		Name: specName,
	})
	if err != nil {
		return fmt.Errorf("listing existing records for %s (%s): %w", specName, spec.Type, err)
	}

	var existing *cf.DNSRecord
	for i := range records {
		record := &records[i]
		if !strings.EqualFold(trimTrailingDot(record.Name), specName) {
			continue
		}
		if normalizeRecordValue(record.Content) == desiredValue && strings.EqualFold(record.Type, spec.Type) {
			return syncMatchingRecord(ctx, api, rc, spec, specName, desiredValue, out, dryRun, dryTable, record)
		}
		existing = record
	}

	if existing != nil {
		return syncExistingRecord(ctx, api, rc, spec, specName, desiredValue, out, dryRun, dryTable, existing)
	}

	// Check for records with same name but different type - we should NOT update these,
	// as multiple record types can coexist for the same name (e.g., A and TXT)
	// Only search for alternative records if we haven't found a matching type already

	if dryRun {
		if dryTable != nil {
			dryTable.AppendRow(table.Row{"ADD", spec.Type, specName, "", spec.Value, "dry-run"})
			return nil
		}
		fmt.Fprintf(out, "would add %s %s -> %s\n", spec.Type, specName, spec.Value)
		return nil
	}

	params := cf.CreateDNSRecordParams{
		Type:    spec.Type,
		Name:    specName,
		Content: spec.Value,
		TTL:     ttlForRecord(spec, nil),
	}
	if spec.Proxied != nil {
		params.Proxied = spec.Proxied
	}

	if _, err := api.CreateDNSRecord(ctx, rc, params); err != nil {
		if utils.IsCloudflareError(err, 81054) || utils.IsCloudflareError(err, 81053) {
			if dryTable != nil {
				dryTable.AppendRow(table.Row{"SKIP", spec.Type, specName, "", spec.Value, "conflicts with existing record"})
				return nil
			}
			fmt.Fprintf(out, "skip add %s %s (conflicts with existing record)\n", spec.Type, specName)
			return nil
		}
		return fmt.Errorf("creating record %s %s -> %s: %w", spec.Type, specName, spec.Value, err)
	}

	if dryTable != nil {
		dryTable.AppendRow(table.Row{"ADD", spec.Type, specName, "", spec.Value, "created"})
	}
	fmt.Fprintf(out, "added %s %s -> %s\n", spec.Type, specName, spec.Value)
	return nil
}

func removeRecord(ctx context.Context, api DNSWriteAPI, rc *cf.ResourceContainer, spec RecordSpec, out io.Writer, dryRun bool, dryTable table.Writer) error {
	specName := trimTrailingDot(spec.Name)
	desiredValue := normalizeRecordValue(spec.Value)

	// Search by name first (more flexible), then filter by type
	records, _, err := api.ListDNSRecords(ctx, rc, cf.ListDNSRecordsParams{
		Name: specName,
	})
	if err != nil {
		return fmt.Errorf("listing existing records for %s (%s): %w", specName, spec.Type, err)
	}

	var matched []cf.DNSRecord
	for _, record := range records {
		recordName := trimTrailingDot(record.Name)
		if !strings.EqualFold(recordName, specName) {
			continue
		}
		if !strings.EqualFold(record.Type, spec.Type) {
			continue
		}
		if spec.Value == "" || normalizeRecordValue(record.Content) == desiredValue {
			matched = append(matched, record)
		}
	}

	if len(matched) == 0 {
		// Try to find any records with this name to show in the table
		var currentValue string
		if allRecords, _, err := api.ListDNSRecords(ctx, rc, cf.ListDNSRecordsParams{Name: specName}); err == nil {
			for _, r := range allRecords {
				if strings.EqualFold(trimTrailingDot(r.Name), specName) && strings.EqualFold(r.Type, spec.Type) {
					currentValue = r.Content
					break
				}
			}
		}
		if dryTable != nil {
			dryTable.AppendRow(table.Row{"SKIP", spec.Type, specName, currentValue, spec.Value, "no match"})
			return nil
		}
		fmt.Fprintf(out, "skip remove %s %s (no match)\n", spec.Type, specName)
		return nil
	}

	if dryRun {
		if dryTable != nil {
			for _, record := range matched {
				dryTable.AppendRow(table.Row{"REMOVE", spec.Type, trimTrailingDot(record.Name), record.Content, spec.Value, "dry-run"})
			}
			return nil
		}
		for _, record := range matched {
			fmt.Fprintf(out, "would remove %s %s -> %s\n", spec.Type, specName, record.Content)
		}
		return nil
	}

	for _, record := range matched {
		if err := api.DeleteDNSRecord(ctx, rc, record.ID); err != nil {
			return fmt.Errorf("deleting record %s %s -> %s: %w", spec.Type, specName, record.Content, err)
		}
		if dryTable != nil {
			dryTable.AppendRow(table.Row{"REMOVE", spec.Type, trimTrailingDot(record.Name), record.Content, spec.Value, "removed"})
		}
		fmt.Fprintf(out, "removed %s %s -> %s\n", spec.Type, specName, record.Content)
	}

	return nil
}

func syncMatchingRecord(ctx context.Context, api DNSWriteAPI, rc *cf.ResourceContainer, spec RecordSpec, specName string, desiredValue string, out io.Writer, dryRun bool, dryTable table.Writer, record *cf.DNSRecord) error {
	desiredTTL := ttlForRecord(spec, record)
	needsTTLUpdate := desiredTTL != record.TTL

	needsProxyUpdate := false
	var proxiedPtr *bool
	if spec.Proxied != nil {
		proxiedPtr = spec.Proxied
		if record.Proxied == nil || *record.Proxied != *spec.Proxied {
			needsProxyUpdate = true
		}
	}

	if !needsTTLUpdate && !needsProxyUpdate {
		if dryTable != nil {
			dryTable.AppendRow(table.Row{"SKIP", spec.Type, specName, record.Content, spec.Value, "already present"})
			return nil
		}
		fmt.Fprintf(out, "skip add %s %s (already present)\n", spec.Type, specName)
		return nil
	}

	if dryRun {
		if dryTable != nil {
			dryTable.AppendRow(table.Row{"UPDATE", spec.Type, specName, record.Content, spec.Value, diffSummary(record, desiredTTL, proxiedPtr)})
			return nil
		}
		fmt.Fprintf(out, "would update %s %s -> %s\n", spec.Type, specName, spec.Value)
		return nil
	}

	params := cf.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    spec.Type,
		Name:    specName,
		Content: spec.Value,
		TTL:     desiredTTL,
	}
	if proxiedPtr != nil {
		params.Proxied = proxiedPtr
	}

	if _, err := api.UpdateDNSRecord(ctx, rc, params); err != nil {
		return fmt.Errorf("updating record %s %s -> %s: %w", spec.Type, specName, spec.Value, err)
	}

	if dryTable != nil {
		dryTable.AppendRow(table.Row{"UPDATE", spec.Type, specName, record.Content, spec.Value, diffSummary(record, desiredTTL, proxiedPtr)})
	}
	fmt.Fprintf(out, "updated %s %s -> %s\n", spec.Type, specName, spec.Value)
	return nil
}

func syncExistingRecord(ctx context.Context, api DNSWriteAPI, rc *cf.ResourceContainer, spec RecordSpec, specName string, desiredValue string, out io.Writer, dryRun bool, dryTable table.Writer, record *cf.DNSRecord) error {
	if dryRun {
		if dryTable != nil {
			dryTable.AppendRow(table.Row{"UPDATE", spec.Type, specName, record.Content, spec.Value, diffSummary(record, ttlForRecord(spec, record), spec.Proxied)})
			return nil
		}
		fmt.Fprintf(out, "would update %s %s -> %s\n", spec.Type, specName, spec.Value)
		return nil
	}

	params := cf.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    spec.Type,
		Name:    specName,
		Content: spec.Value,
		TTL:     ttlForRecord(spec, record),
	}
	if spec.Proxied != nil {
		params.Proxied = spec.Proxied
	}

	if _, err := api.UpdateDNSRecord(ctx, rc, params); err != nil {
		return fmt.Errorf("updating record %s %s -> %s: %w", spec.Type, specName, spec.Value, err)
	}

	if dryTable != nil {
		dryTable.AppendRow(table.Row{"UPDATE", spec.Type, specName, record.Content, spec.Value, diffSummary(record, ttlForRecord(spec, record), spec.Proxied)})
	}
	fmt.Fprintf(out, "updated %s %s -> %s\n", spec.Type, specName, spec.Value)
	return nil
}

func ttlForRecord(spec RecordSpec, record *cf.DNSRecord) int {
	if spec.TTL != nil && *spec.TTL > 0 {
		return *spec.TTL
	}
	if record != nil && record.TTL > 0 {
		return record.TTL
	}
	return 1
}

func diffSummary(record *cf.DNSRecord, desiredTTL int, desiredProxied *bool) string {
	currentProxied := ""
	if record.Proxied != nil {
		currentProxied = fmt.Sprintf("proxied=%t", *record.Proxied)
	}
	desiredProxiedStr := ""
	if desiredProxied != nil {
		desiredProxiedStr = fmt.Sprintf("->%t", *desiredProxied)
	}
	return fmt.Sprintf("ttl=%d->%d %s%s", record.TTL, desiredTTL, currentProxied, desiredProxiedStr)
}

func trimTrailingDot(name string) string {
	return strings.TrimSuffix(name, ".")
}

func normalizeRecordValue(val string) string {
	val = strings.TrimSpace(val)
	// Strip surrounding quotes (common in TXT records from Cloudflare)
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = val[1 : len(val)-1]
	}
	return trimTrailingDot(val)
}
