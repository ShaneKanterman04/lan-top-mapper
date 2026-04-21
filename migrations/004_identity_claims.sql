ALTER TABLE devices ADD COLUMN name_source TEXT;
ALTER TABLE devices ADD COLUMN vendor_source TEXT;
ALTER TABLE devices ADD COLUMN type_source TEXT;

ALTER TABLE device_observations ADD COLUMN name_source TEXT;
ALTER TABLE device_observations ADD COLUMN vendor_source TEXT;
ALTER TABLE device_observations ADD COLUMN type_source TEXT;

CREATE TABLE IF NOT EXISTS observation_identity_claims (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_observation_id INTEGER NOT NULL,
    source TEXT NOT NULL,
    claim_type TEXT NOT NULL,
    value TEXT NOT NULL,
    confidence_label TEXT,
    raw_json TEXT,
    FOREIGN KEY(device_observation_id) REFERENCES device_observations(id)
);

CREATE INDEX IF NOT EXISTS idx_observation_identity_claims_observation ON observation_identity_claims(device_observation_id);
CREATE INDEX IF NOT EXISTS idx_observation_identity_claims_source ON observation_identity_claims(source);
CREATE INDEX IF NOT EXISTS idx_observation_identity_claims_type ON observation_identity_claims(claim_type);
