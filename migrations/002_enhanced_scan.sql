ALTER TABLE scan_runs ADD COLUMN scan_mode TEXT NOT NULL DEFAULT 'limited';
ALTER TABLE scan_runs ADD COLUMN total_targets INTEGER NOT NULL DEFAULT 0;
ALTER TABLE scan_runs ADD COLUMN devices_found INTEGER NOT NULL DEFAULT 0;
ALTER TABLE scan_runs ADD COLUMN duration_ms INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS observation_evidence (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_observation_id INTEGER NOT NULL,
    source TEXT NOT NULL,
    FOREIGN KEY(device_observation_id) REFERENCES device_observations(id)
);

CREATE INDEX IF NOT EXISTS idx_observation_evidence_observation ON observation_evidence(device_observation_id);
CREATE INDEX IF NOT EXISTS idx_observation_evidence_source ON observation_evidence(source);
