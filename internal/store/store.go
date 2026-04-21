package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"lan-topology-mapper/internal/models"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}

	store := &Store{db: db}
	if err := store.runMigrations(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) runMigrations(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at DATETIME NOT NULL)`); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	entries, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		applied, err := s.hasMigration(ctx, entry.Name())
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		raw, err := os.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration transaction: %w", err)
		}
		if _, err := tx.ExecContext(ctx, string(raw)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)`, entry.Name(), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (s *Store) hasMigration(ctx context.Context, name string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE name = ?`, name).Scan(&count); err != nil {
		return false, fmt.Errorf("check migration %s: %w", name, err)
	}
	return count > 0, nil
}

func (s *Store) CreateScanRun(ctx context.Context, subnet, gatewayIP string, diagnostics models.ScanDiagnostics) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO scan_runs (subnet, gateway_ip, status, started_at, scan_mode, total_targets, devices_found, duration_ms) VALUES (?, ?, 'running', ?, ?, ?, 0, 0)`,
		subnet,
		gatewayIP,
		now,
		diagnostics.Mode,
		diagnostics.TotalTargets,
	)
	if err != nil {
		return 0, fmt.Errorf("insert scan run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read scan run id: %w", err)
	}
	return id, nil
}

func (s *Store) CompleteScanRun(ctx context.Context, scanID int64, diagnostics models.ScanDiagnostics) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE scan_runs SET status = 'completed', completed_at = ?, scan_mode = ?, total_targets = ?, devices_found = ?, duration_ms = ? WHERE id = ?`,
		now,
		diagnostics.Mode,
		diagnostics.TotalTargets,
		diagnostics.DevicesFound,
		diagnostics.DurationMs,
		scanID,
	)
	if err != nil {
		return fmt.Errorf("complete scan run: %w", err)
	}
	return nil
}

func (s *Store) FailScanRun(ctx context.Context, scanID int64, failure error) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE scan_runs SET status = 'failed', completed_at = ?, error_message = ? WHERE id = ?`,
		now,
		failure.Error(),
		scanID,
	)
	if err != nil {
		return fmt.Errorf("fail scan run: %w", err)
	}
	return nil
}

func (s *Store) ReplaceScanObservations(ctx context.Context, scanID int64, observations []models.DeviceObservation) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `UPDATE devices SET is_online = 0`); err != nil {
		return fmt.Errorf("mark devices offline: %w", err)
	}

	for _, observation := range observations {
		deviceID, err := upsertDevice(ctx, tx, observation, now)
		if err != nil {
			return err
		}

		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO device_observations (scan_run_id, device_id, ip_address, mac_address, hostname, vendor, display_name, name_source, device_type, type_source, classification_confidence, classification_reasons, vendor_source, is_online, observed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			scanID,
			deviceID,
			observation.IP,
			nullable(observation.MAC),
			nullable(observation.Hostname),
			nullable(observation.Vendor),
			nullable(observation.DisplayName),
			nullable(observation.NameSource),
			nullable(observation.DeviceType),
			nullable(observation.TypeSource),
			nullable(observation.ClassificationConfidence),
			nullable(mustJSON(observation.ClassificationReasons)),
			nullable(observation.VendorSource),
			observation.Online,
			now,
		)
		if err != nil {
			return fmt.Errorf("insert observation for %s: %w", observation.IP, err)
		}

		observationID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("read observation id for %s: %w", observation.IP, err)
		}

		for _, service := range observation.Services {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO services (device_observation_id, port, protocol, service_name) VALUES (?, ?, ?, ?)`,
				observationID,
				service.Port,
				service.Protocol,
				service.ServiceName,
			); err != nil {
				return fmt.Errorf("insert service for %s: %w", observation.IP, err)
			}
		}

		for _, source := range observation.Evidence {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO observation_evidence (device_observation_id, source) VALUES (?, ?)`,
				observationID,
				source,
			); err != nil {
				return fmt.Errorf("insert evidence for %s: %w", observation.IP, err)
			}
		}

		for _, claim := range observation.IdentityClaims {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO observation_identity_claims (device_observation_id, source, claim_type, value, confidence_label, raw_json) VALUES (?, ?, ?, ?, ?, ?)`,
				observationID,
				claim.Source,
				claim.ClaimType,
				claim.Value,
				nullable(claim.ConfidenceLabel),
				nullable(claim.RawJSON),
			); err != nil {
				return fmt.Errorf("insert identity claim for %s: %w", observation.IP, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit scan transaction: %w", err)
	}
	return nil
}

func upsertDevice(ctx context.Context, tx *sql.Tx, observation models.DeviceObservation, now string) (int64, error) {
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO devices (ip_address, mac_address, hostname, vendor, display_name, name_source, device_type, type_source, classification_confidence, classification_reasons, vendor_source, first_seen, last_seen, is_online)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(ip_address) DO UPDATE SET
		   mac_address = COALESCE(excluded.mac_address, devices.mac_address),
		   hostname = COALESCE(excluded.hostname, devices.hostname),
		   vendor = COALESCE(excluded.vendor, devices.vendor),
		   display_name = COALESCE(excluded.display_name, devices.display_name),
		   name_source = COALESCE(excluded.name_source, devices.name_source),
		   device_type = COALESCE(excluded.device_type, devices.device_type),
		   type_source = COALESCE(excluded.type_source, devices.type_source),
		   classification_confidence = COALESCE(excluded.classification_confidence, devices.classification_confidence),
		   classification_reasons = COALESCE(excluded.classification_reasons, devices.classification_reasons),
		   vendor_source = COALESCE(excluded.vendor_source, devices.vendor_source),
		   last_seen = excluded.last_seen,
		   is_online = excluded.is_online`,
		observation.IP,
		nullable(observation.MAC),
		nullable(observation.Hostname),
		nullable(observation.Vendor),
		nullable(observation.DisplayName),
		nullable(observation.NameSource),
		nullable(observation.DeviceType),
		nullable(observation.TypeSource),
		nullable(observation.ClassificationConfidence),
		nullable(mustJSON(observation.ClassificationReasons)),
		nullable(observation.VendorSource),
		now,
		now,
		observation.Online,
	); err != nil {
		return 0, fmt.Errorf("upsert device %s: %w", observation.IP, err)
	}

	var deviceID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM devices WHERE ip_address = ?`, observation.IP).Scan(&deviceID); err != nil {
		return 0, fmt.Errorf("lookup device id for %s: %w", observation.IP, err)
	}
	return deviceID, nil
}

func (s *Store) ListDevices(ctx context.Context) ([]models.Device, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, ip_address, COALESCE(mac_address, ''), COALESCE(hostname, ''), COALESCE(vendor, ''), COALESCE(display_name, ''), COALESCE(name_source, ''), COALESCE(device_type, ''), COALESCE(type_source, ''), COALESCE(classification_confidence, ''), COALESCE(classification_reasons, ''), COALESCE(vendor_source, ''), first_seen, last_seen, is_online FROM devices ORDER BY ip_address`)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	devices := make([]models.Device, 0)
	for rows.Next() {
		var device models.Device
		var firstSeen string
		var lastSeen string
		var classificationReasons string
		if err := rows.Scan(&device.ID, &device.IP, &device.MAC, &device.Hostname, &device.Vendor, &device.DisplayName, &device.NameSource, &device.DeviceType, &device.TypeSource, &device.ClassificationConfidence, &classificationReasons, &device.VendorSource, &firstSeen, &lastSeen, &device.Online); err != nil {
			return nil, fmt.Errorf("scan device row: %w", err)
		}
		device.ClassificationReasons = parseJSONArray(classificationReasons)
		device.FirstSeen, _ = time.Parse(time.RFC3339Nano, firstSeen)
		device.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeen)
		device.Services, _ = s.listLatestServicesForDevice(ctx, device.ID)
		device.Evidence, _ = s.listLatestEvidenceForDevice(ctx, device.ID)
		device.IdentityClaims, _ = s.listLatestIdentityClaimsForDevice(ctx, device.ID)
		devices = append(devices, device)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device rows: %w", err)
	}

	return devices, nil
}

func (s *Store) GetDevice(ctx context.Context, id int64) (*models.Device, error) {
	return s.fetchDevice(ctx, id)
}

func (s *Store) GetDeviceDetail(ctx context.Context, id int64) (*models.DeviceDetail, error) {
	device, err := s.fetchDevice(ctx, id)
	if err != nil || device == nil {
		return nil, err
	}

	history, err := s.listDeviceObservationHistory(ctx, id, 20)
	if err != nil {
		return nil, err
	}
	evidenceSummary, err := s.deviceEvidenceSummary(ctx, id)
	if err != nil {
		return nil, err
	}
	scanSummary, err := s.deviceScanSummary(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &models.DeviceDetail{
		Device:             *device,
		ObservationHistory: history,
		EvidenceSummary:    evidenceSummary,
		ScanSummary:        scanSummary,
	}
	if len(history) > 0 {
		detail.LatestObservation = &history[0]
	}
	return detail, nil
}

func (s *Store) fetchDevice(ctx context.Context, id int64) (*models.Device, error) {
	var device models.Device
	var firstSeen string
	var lastSeen string
	var classificationReasons string
	err := s.db.QueryRowContext(ctx, `SELECT id, ip_address, COALESCE(mac_address, ''), COALESCE(hostname, ''), COALESCE(vendor, ''), COALESCE(display_name, ''), COALESCE(name_source, ''), COALESCE(device_type, ''), COALESCE(type_source, ''), COALESCE(classification_confidence, ''), COALESCE(classification_reasons, ''), COALESCE(vendor_source, ''), first_seen, last_seen, is_online FROM devices WHERE id = ?`, id).Scan(
		&device.ID,
		&device.IP,
		&device.MAC,
		&device.Hostname,
		&device.Vendor,
		&device.DisplayName,
		&device.NameSource,
		&device.DeviceType,
		&device.TypeSource,
		&device.ClassificationConfidence,
		&classificationReasons,
		&device.VendorSource,
		&firstSeen,
		&lastSeen,
		&device.Online,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	device.ClassificationReasons = parseJSONArray(classificationReasons)
	device.FirstSeen, _ = time.Parse(time.RFC3339Nano, firstSeen)
	device.LastSeen, _ = time.Parse(time.RFC3339Nano, lastSeen)
	device.Services, _ = s.listLatestServicesForDevice(ctx, id)
	device.Evidence, _ = s.listLatestEvidenceForDevice(ctx, id)
	device.IdentityClaims, _ = s.listLatestIdentityClaimsForDevice(ctx, id)
	return &device, nil
}

func (s *Store) listLatestIdentityClaimsForDevice(ctx context.Context, deviceID int64) ([]models.IdentityClaim, error) {
	var observationID int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM device_observations WHERE device_id = ? ORDER BY observed_at DESC, id DESC LIMIT 1`, deviceID).Scan(&observationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup latest observation for identity claims: %w", err)
	}
	return s.listIdentityClaimsForObservation(ctx, observationID)
}

func (s *Store) listDeviceObservationHistory(ctx context.Context, deviceID int64, limit int) ([]models.DeviceObservationDetail, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT o.id, o.scan_run_id, COALESCE(r.status, ''), COALESCE(r.started_at, ''), o.observed_at,
		       o.ip_address, COALESCE(o.mac_address, ''), COALESCE(o.hostname, ''), COALESCE(o.vendor, ''),
		       COALESCE(o.display_name, ''), COALESCE(o.name_source, ''), COALESCE(o.device_type, ''), COALESCE(o.type_source, ''), COALESCE(o.classification_confidence, ''), COALESCE(o.classification_reasons, ''), COALESCE(o.vendor_source, ''), o.is_online
		FROM device_observations o
		LEFT JOIN scan_runs r ON r.id = o.scan_run_id
		WHERE o.device_id = ?
		ORDER BY o.observed_at DESC, o.id DESC
		LIMIT ?`, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("list device observation history: %w", err)
	}
	defer rows.Close()

	history := make([]models.DeviceObservationDetail, 0)
	for rows.Next() {
		var item models.DeviceObservationDetail
		var scanStartedAt string
		var observedAt string
		var classificationReasons string
		if err := rows.Scan(
			&item.ObservationID,
			&item.ScanID,
			&item.ScanStatus,
			&scanStartedAt,
			&observedAt,
			&item.IP,
			&item.MAC,
			&item.Hostname,
			&item.Vendor,
			&item.DisplayName,
			&item.NameSource,
			&item.DeviceType,
			&item.TypeSource,
			&item.ClassificationConfidence,
			&classificationReasons,
			&item.VendorSource,
			&item.Online,
		); err != nil {
			return nil, fmt.Errorf("scan device observation history row: %w", err)
		}
		if parsed, err := time.Parse(time.RFC3339Nano, observedAt); err == nil {
			item.ObservedAt = parsed
		}
		if scanStartedAt != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, scanStartedAt); err == nil {
				item.ScanStartedAt = &parsed
			}
		}
		item.ClassificationReasons = parseJSONArray(classificationReasons)
		item.Services, _ = s.listServicesForObservation(ctx, item.ObservationID)
		item.Evidence, _ = s.listEvidenceForObservation(ctx, item.ObservationID)
		item.IdentityClaims, _ = s.listIdentityClaimsForObservation(ctx, item.ObservationID)
		history = append(history, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device observation history rows: %w", err)
	}
	return history, nil
}

func (s *Store) listServicesForObservation(ctx context.Context, observationID int64) ([]models.Service, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT port, protocol, service_name FROM services WHERE device_observation_id = ? ORDER BY port`, observationID)
	if err != nil {
		return nil, fmt.Errorf("list services for observation: %w", err)
	}
	defer rows.Close()

	services := make([]models.Service, 0)
	for rows.Next() {
		var service models.Service
		if err := rows.Scan(&service.Port, &service.Protocol, &service.ServiceName); err != nil {
			return nil, fmt.Errorf("scan service observation row: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate service observation rows: %w", err)
	}
	return services, nil
}

func (s *Store) listEvidenceForObservation(ctx context.Context, observationID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT source FROM observation_evidence WHERE device_observation_id = ? ORDER BY source`, observationID)
	if err != nil {
		return nil, fmt.Errorf("list evidence for observation: %w", err)
	}
	defer rows.Close()

	evidence := make([]string, 0)
	for rows.Next() {
		var source string
		if err := rows.Scan(&source); err != nil {
			return nil, fmt.Errorf("scan evidence observation row: %w", err)
		}
		evidence = append(evidence, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidence observation rows: %w", err)
	}
	return evidence, nil
}

func (s *Store) listIdentityClaimsForObservation(ctx context.Context, observationID int64) ([]models.IdentityClaim, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT source, claim_type, value, COALESCE(confidence_label, ''), COALESCE(raw_json, '') FROM observation_identity_claims WHERE device_observation_id = ? ORDER BY source, claim_type, value`, observationID)
	if err != nil {
		return nil, fmt.Errorf("list identity claims for observation: %w", err)
	}
	defer rows.Close()

	claims := make([]models.IdentityClaim, 0)
	for rows.Next() {
		var claim models.IdentityClaim
		if err := rows.Scan(&claim.Source, &claim.ClaimType, &claim.Value, &claim.ConfidenceLabel, &claim.RawJSON); err != nil {
			return nil, fmt.Errorf("scan identity claim row: %w", err)
		}
		claims = append(claims, claim)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity claim rows: %w", err)
	}
	return claims, nil
}

func (s *Store) deviceEvidenceSummary(ctx context.Context, deviceID int64) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.source, COUNT(1)
		FROM observation_evidence e
		JOIN device_observations o ON o.id = e.device_observation_id
		WHERE o.device_id = ?
		GROUP BY e.source
		ORDER BY e.source`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("device evidence summary: %w", err)
	}
	defer rows.Close()

	summary := map[string]int{}
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("scan device evidence summary row: %w", err)
		}
		summary[source] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device evidence summary rows: %w", err)
	}
	return summary, nil
}

func (s *Store) deviceScanSummary(ctx context.Context, deviceID int64) (models.DeviceScanSummary, error) {
	var summary models.DeviceScanSummary
	var firstSeen sql.NullString
	var lastSeen sql.NullString
	var lastSeenScanID sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1), MIN(observed_at), MAX(observed_at),
		       (SELECT scan_run_id FROM device_observations WHERE device_id = ? ORDER BY observed_at DESC, id DESC LIMIT 1)
		FROM device_observations
		WHERE device_id = ?`, deviceID, deviceID).Scan(&summary.TimesSeen, &firstSeen, &lastSeen, &lastSeenScanID)
	if err != nil {
		return models.DeviceScanSummary{}, fmt.Errorf("device scan summary: %w", err)
	}
	if firstSeen.Valid {
		if parsed, err := time.Parse(time.RFC3339Nano, firstSeen.String); err == nil {
			summary.FirstSeenAt = &parsed
		}
	}
	if lastSeen.Valid {
		if parsed, err := time.Parse(time.RFC3339Nano, lastSeen.String); err == nil {
			summary.LastSeenAt = &parsed
		}
	}
	if lastSeenScanID.Valid {
		id := lastSeenScanID.Int64
		summary.LastSeenScanID = &id
	}
	return summary, nil
}

func (s *Store) listLatestServicesForDevice(ctx context.Context, deviceID int64) ([]models.Service, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.port, s.protocol, s.service_name
		FROM services s
		JOIN device_observations o ON o.id = s.device_observation_id
		WHERE o.id = (
			SELECT id FROM device_observations WHERE device_id = ? ORDER BY observed_at DESC, id DESC LIMIT 1
		)
		ORDER BY s.port`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	services := make([]models.Service, 0)
	for rows.Next() {
		var service models.Service
		if err := rows.Scan(&service.Port, &service.Protocol, &service.ServiceName); err != nil {
			return nil, fmt.Errorf("scan service row: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate service rows: %w", err)
	}
	return services, nil
}

func (s *Store) listLatestEvidenceForDevice(ctx context.Context, deviceID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT e.source
		FROM observation_evidence e
		JOIN device_observations o ON o.id = e.device_observation_id
		WHERE o.id = (
			SELECT id FROM device_observations WHERE device_id = ? ORDER BY observed_at DESC, id DESC LIMIT 1
		)
		ORDER BY e.source`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list evidence: %w", err)
	}
	defer rows.Close()

	evidence := make([]string, 0)
	for rows.Next() {
		var source string
		if err := rows.Scan(&source); err != nil {
			return nil, fmt.Errorf("scan evidence row: %w", err)
		}
		evidence = append(evidence, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidence rows: %w", err)
	}
	return evidence, nil
}

func (s *Store) GetScanRun(ctx context.Context, id int64) (*models.ScanRun, error) {
	var scan models.ScanRun
	var startedAt string
	var completedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id, subnet, COALESCE(gateway_ip, ''), status, started_at, completed_at, COALESCE(error_message, ''), scan_mode, total_targets, devices_found, duration_ms FROM scan_runs WHERE id = ?`, id).Scan(
		&scan.ID,
		&scan.Subnet,
		&scan.GatewayIP,
		&scan.Status,
		&startedAt,
		&completedAt,
		&scan.Error,
		&scan.Diagnostics.Mode,
		&scan.Diagnostics.TotalTargets,
		&scan.Diagnostics.DevicesFound,
		&scan.Diagnostics.DurationMs,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get scan run: %w", err)
	}
	if parsed, err := time.Parse(time.RFC3339Nano, startedAt); err == nil {
		scan.StartedAt = parsed
	}
	if completedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err == nil {
			scan.CompletedAt = &parsed
		}
	}
	scan.Diagnostics.EvidenceCounts, _ = s.scanEvidenceCounts(ctx, id)
	return &scan, nil
}

func (s *Store) LatestScanRun(ctx context.Context) (*models.ScanRun, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, subnet, COALESCE(gateway_ip, ''), status, started_at, completed_at, COALESCE(error_message, ''), scan_mode, total_targets, devices_found, duration_ms FROM scan_runs ORDER BY started_at DESC, id DESC LIMIT 1`)
	var scan models.ScanRun
	var startedAt string
	var completedAt sql.NullString
	if err := row.Scan(&scan.ID, &scan.Subnet, &scan.GatewayIP, &scan.Status, &startedAt, &completedAt, &scan.Error, &scan.Diagnostics.Mode, &scan.Diagnostics.TotalTargets, &scan.Diagnostics.DevicesFound, &scan.Diagnostics.DurationMs); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("latest scan run: %w", err)
	}
	if parsed, err := time.Parse(time.RFC3339Nano, startedAt); err == nil {
		scan.StartedAt = parsed
	}
	if completedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err == nil {
			scan.CompletedAt = &parsed
		}
	}
	scan.Diagnostics.EvidenceCounts, _ = s.scanEvidenceCounts(ctx, scan.ID)
	return &scan, nil
}

func (s *Store) scanEvidenceCounts(ctx context.Context, scanID int64) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT e.source, COUNT(1)
		FROM observation_evidence e
		JOIN device_observations o ON o.id = e.device_observation_id
		WHERE o.scan_run_id = ?
		GROUP BY e.source`, scanID)
	if err != nil {
		return nil, fmt.Errorf("scan evidence counts: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("scan evidence count row: %w", err)
		}
		counts[source] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scan evidence counts: %w", err)
	}
	return counts, nil
}

func (s *Store) Summary(ctx context.Context, capabilities models.ScanDiagnostics) (models.Summary, error) {
	devices, err := s.ListDevices(ctx)
	if err != nil {
		return models.Summary{}, err
	}

	latestScan, err := s.LatestScanRun(ctx)
	if err != nil {
		return models.Summary{}, err
	}

	summary := models.Summary{
		DeviceCount:      len(devices),
		LatestScanState:  "idle",
		LimitedMode:      capabilities.Mode != "enhanced",
		CurrentMode:      capabilities.Mode,
		AvailableSources: append([]string(nil), capabilities.AvailableSources...),
	}

	for _, device := range devices {
		if device.Online {
			summary.OnlineCount++
		} else {
			summary.OfflineCount++
		}
	}

	if latestScan != nil {
		summary.Subnet = latestScan.Subnet
		summary.GatewayIP = latestScan.GatewayIP
		summary.LatestScanState = latestScan.Status
		if latestScan.CompletedAt != nil {
			summary.LastScanAt = latestScan.CompletedAt
		}
		scanID := latestScan.ID
		summary.LatestScanID = &scanID
		diagnostics := latestScan.Diagnostics
		diagnostics.AvailableSources = append([]string(nil), summary.AvailableSources...)
		summary.LatestScanDiagnostics = &diagnostics
	}

	return summary, nil
}

func (s *Store) Topology(ctx context.Context) (models.Topology, error) {
	devices, err := s.ListDevices(ctx)
	if err != nil {
		return models.Topology{}, err
	}
	latestScan, err := s.LatestScanRun(ctx)
	if err != nil {
		return models.Topology{}, err
	}

	topology := models.Topology{Nodes: []models.TopologyNode{}, Edges: []models.TopologyEdge{}}
	if latestScan == nil {
		return topology, nil
	}

	gatewayID := "gateway"
	var gatewayDevice *models.Device
	filteredDevices := make([]models.Device, 0, len(devices))
	for _, device := range devices {
		if device.IP == latestScan.GatewayIP && gatewayDevice == nil {
			deviceCopy := device
			gatewayDevice = &deviceCopy
			continue
		}
		filteredDevices = append(filteredDevices, device)
	}

	gatewayData := map[string]any{
		"id":    gatewayID,
		"label": firstNonEmpty(latestScan.GatewayIP, "Gateway"),
		"kind":  "gateway",
		"ip":    latestScan.GatewayIP,
	}
	if gatewayDevice != nil {
		gatewayData["deviceId"] = gatewayDevice.ID
		gatewayData["label"] = firstNonEmpty(gatewayDevice.DisplayName, gatewayDevice.Hostname, gatewayDevice.IP, "Gateway")
		gatewayData["hostname"] = gatewayDevice.Hostname
		gatewayData["mac"] = gatewayDevice.MAC
		gatewayData["vendor"] = gatewayDevice.Vendor
		gatewayData["displayName"] = gatewayDevice.DisplayName
		gatewayData["nameSource"] = gatewayDevice.NameSource
		gatewayData["deviceType"] = gatewayDevice.DeviceType
		gatewayData["typeSource"] = gatewayDevice.TypeSource
		gatewayData["classificationConfidence"] = gatewayDevice.ClassificationConfidence
		gatewayData["classificationReasons"] = gatewayDevice.ClassificationReasons
		gatewayData["vendorSource"] = gatewayDevice.VendorSource
		gatewayData["identityClaims"] = gatewayDevice.IdentityClaims
		gatewayData["online"] = gatewayDevice.Online
		gatewayData["services"] = gatewayDevice.Services
		gatewayData["evidence"] = gatewayDevice.Evidence
	}
	topology.Nodes = append(topology.Nodes, models.TopologyNode{Data: map[string]any{
		"id":                       gatewayData["id"],
		"deviceId":                 gatewayData["deviceId"],
		"label":                    gatewayData["label"],
		"kind":                     gatewayData["kind"],
		"ip":                       gatewayData["ip"],
		"hostname":                 gatewayData["hostname"],
		"mac":                      gatewayData["mac"],
		"vendor":                   gatewayData["vendor"],
		"displayName":              gatewayData["displayName"],
		"nameSource":               gatewayData["nameSource"],
		"deviceType":               gatewayData["deviceType"],
		"typeSource":               gatewayData["typeSource"],
		"classificationConfidence": gatewayData["classificationConfidence"],
		"classificationReasons":    gatewayData["classificationReasons"],
		"vendorSource":             gatewayData["vendorSource"],
		"identityClaims":           gatewayData["identityClaims"],
		"online":                   gatewayData["online"],
		"services":                 gatewayData["services"],
		"evidence":                 gatewayData["evidence"],
	}})

	for _, device := range filteredDevices {
		nodeID := fmt.Sprintf("device-%d", device.ID)
		label := firstNonEmpty(device.DisplayName, device.Hostname, device.IP)
		topology.Nodes = append(topology.Nodes, models.TopologyNode{Data: map[string]any{
			"id":                       nodeID,
			"deviceId":                 device.ID,
			"label":                    label,
			"ip":                       device.IP,
			"hostname":                 device.Hostname,
			"mac":                      device.MAC,
			"vendor":                   device.Vendor,
			"displayName":              device.DisplayName,
			"nameSource":               device.NameSource,
			"deviceType":               device.DeviceType,
			"typeSource":               device.TypeSource,
			"classificationConfidence": device.ClassificationConfidence,
			"classificationReasons":    device.ClassificationReasons,
			"vendorSource":             device.VendorSource,
			"identityClaims":           device.IdentityClaims,
			"online":                   device.Online,
			"services":                 device.Services,
			"evidence":                 device.Evidence,
			"kind":                     "device",
		}})
		topology.Edges = append(topology.Edges, models.TopologyEdge{Data: map[string]any{
			"id":     fmt.Sprintf("edge-%d", device.ID),
			"source": gatewayID,
			"target": nodeID,
		}})
	}

	return topology, nil
}

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func mustJSON(values []string) string {
	if len(values) == 0 {
		return ""
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return string(raw)
}

func parseJSONArray(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return values
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
