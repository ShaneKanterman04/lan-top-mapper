CREATE TABLE IF NOT EXISTS scan_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subnet TEXT NOT NULL,
    gateway_ip TEXT,
    status TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    error_message TEXT
);

CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT NOT NULL UNIQUE,
    mac_address TEXT,
    hostname TEXT,
    vendor TEXT,
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    is_online BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS device_observations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_run_id INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    ip_address TEXT NOT NULL,
    mac_address TEXT,
    hostname TEXT,
    vendor TEXT,
    is_online BOOLEAN NOT NULL,
    observed_at DATETIME NOT NULL,
    FOREIGN KEY(scan_run_id) REFERENCES scan_runs(id),
    FOREIGN KEY(device_id) REFERENCES devices(id)
);

CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_observation_id INTEGER NOT NULL,
    port INTEGER NOT NULL,
    protocol TEXT NOT NULL,
    service_name TEXT NOT NULL,
    FOREIGN KEY(device_observation_id) REFERENCES device_observations(id)
);

CREATE INDEX IF NOT EXISTS idx_device_observations_scan_run ON device_observations(scan_run_id);
CREATE INDEX IF NOT EXISTS idx_services_observation ON services(device_observation_id);
