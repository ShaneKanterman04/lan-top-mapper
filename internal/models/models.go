package models

import "time"

type Service struct {
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ServiceName string `json:"serviceName"`
}

type IdentityClaim struct {
	Source          string `json:"source"`
	ClaimType       string `json:"claimType"`
	Value           string `json:"value"`
	ConfidenceLabel string `json:"confidenceLabel,omitempty"`
	RawJSON         string `json:"rawJson,omitempty"`
}

type ScanDiagnostics struct {
	Mode             string         `json:"mode"`
	AvailableSources []string       `json:"availableSources,omitempty"`
	EvidenceCounts   map[string]int `json:"evidenceCounts,omitempty"`
	TotalTargets     int            `json:"totalTargets"`
	DevicesFound     int            `json:"devicesFound"`
	DurationMs       int64          `json:"durationMs"`
}

type DeviceObservation struct {
	IP                       string          `json:"ip"`
	MAC                      string          `json:"mac,omitempty"`
	Hostname                 string          `json:"hostname,omitempty"`
	Vendor                   string          `json:"vendor,omitempty"`
	DisplayName              string          `json:"displayName,omitempty"`
	NameSource               string          `json:"nameSource,omitempty"`
	DeviceType               string          `json:"deviceType,omitempty"`
	TypeSource               string          `json:"typeSource,omitempty"`
	ClassificationConfidence string          `json:"classificationConfidence,omitempty"`
	ClassificationReasons    []string        `json:"classificationReasons,omitempty"`
	VendorSource             string          `json:"vendorSource,omitempty"`
	IdentityClaims           []IdentityClaim `json:"identityClaims,omitempty"`
	Online                   bool            `json:"online"`
	Services                 []Service       `json:"services,omitempty"`
	Evidence                 []string        `json:"evidence,omitempty"`
}

type ScanTarget struct {
	IP     string
	MAC    string
	Source string
	Claims []IdentityClaim
}

type Device struct {
	ID                       int64           `json:"id"`
	IP                       string          `json:"ip"`
	MAC                      string          `json:"mac,omitempty"`
	Hostname                 string          `json:"hostname,omitempty"`
	Vendor                   string          `json:"vendor,omitempty"`
	DisplayName              string          `json:"displayName,omitempty"`
	NameSource               string          `json:"nameSource,omitempty"`
	DeviceType               string          `json:"deviceType,omitempty"`
	TypeSource               string          `json:"typeSource,omitempty"`
	ClassificationConfidence string          `json:"classificationConfidence,omitempty"`
	ClassificationReasons    []string        `json:"classificationReasons,omitempty"`
	VendorSource             string          `json:"vendorSource,omitempty"`
	IdentityClaims           []IdentityClaim `json:"identityClaims,omitempty"`
	Online                   bool            `json:"online"`
	FirstSeen                time.Time       `json:"firstSeen"`
	LastSeen                 time.Time       `json:"lastSeen"`
	Services                 []Service       `json:"services,omitempty"`
	Evidence                 []string        `json:"evidence,omitempty"`
}

type DeviceDetail struct {
	Device             Device                    `json:"device"`
	LatestObservation  *DeviceObservationDetail  `json:"latestObservation,omitempty"`
	ObservationHistory []DeviceObservationDetail `json:"observationHistory,omitempty"`
	EvidenceSummary    map[string]int            `json:"evidenceSummary,omitempty"`
	ScanSummary        DeviceScanSummary         `json:"scanSummary"`
}

type DeviceObservationDetail struct {
	ObservationID            int64           `json:"observationId"`
	ScanID                   int64           `json:"scanId"`
	ScanStatus               string          `json:"scanStatus"`
	ScanStartedAt            *time.Time      `json:"scanStartedAt,omitempty"`
	ObservedAt               time.Time       `json:"observedAt"`
	IP                       string          `json:"ip"`
	MAC                      string          `json:"mac,omitempty"`
	Hostname                 string          `json:"hostname,omitempty"`
	Vendor                   string          `json:"vendor,omitempty"`
	DisplayName              string          `json:"displayName,omitempty"`
	NameSource               string          `json:"nameSource,omitempty"`
	DeviceType               string          `json:"deviceType,omitempty"`
	TypeSource               string          `json:"typeSource,omitempty"`
	ClassificationConfidence string          `json:"classificationConfidence,omitempty"`
	ClassificationReasons    []string        `json:"classificationReasons,omitempty"`
	VendorSource             string          `json:"vendorSource,omitempty"`
	IdentityClaims           []IdentityClaim `json:"identityClaims,omitempty"`
	Online                   bool            `json:"online"`
	Evidence                 []string        `json:"evidence,omitempty"`
	Services                 []Service       `json:"services,omitempty"`
}

type DeviceScanSummary struct {
	TimesSeen      int        `json:"timesSeen"`
	FirstSeenAt    *time.Time `json:"firstSeenAt,omitempty"`
	LastSeenAt     *time.Time `json:"lastSeenAt,omitempty"`
	LastSeenScanID *int64     `json:"lastSeenScanId,omitempty"`
}

type ScanRun struct {
	ID          int64           `json:"id"`
	Subnet      string          `json:"subnet"`
	GatewayIP   string          `json:"gatewayIp,omitempty"`
	Status      string          `json:"status"`
	StartedAt   time.Time       `json:"startedAt"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	Error       string          `json:"error,omitempty"`
	Diagnostics ScanDiagnostics `json:"diagnostics"`
}

type Summary struct {
	Subnet                string           `json:"subnet"`
	GatewayIP             string           `json:"gatewayIp,omitempty"`
	DeviceCount           int              `json:"deviceCount"`
	OnlineCount           int              `json:"onlineCount"`
	OfflineCount          int              `json:"offlineCount"`
	LastScanAt            *time.Time       `json:"lastScanAt,omitempty"`
	LatestScanID          *int64           `json:"latestScanId,omitempty"`
	LatestScanState       string           `json:"latestScanState"`
	LimitedMode           bool             `json:"limitedMode"`
	CurrentMode           string           `json:"currentMode"`
	AvailableSources      []string         `json:"availableSources,omitempty"`
	LatestScanDiagnostics *ScanDiagnostics `json:"latestScanDiagnostics,omitempty"`
}

type TopologyNode struct {
	Data map[string]any `json:"data"`
}

type TopologyEdge struct {
	Data map[string]any `json:"data"`
}

type Topology struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}
