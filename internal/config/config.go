package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr              string
	DBPath            string
	StaticDir         string
	SubnetOverride    string
	Concurrency       int
	ConnectTimeout    time.Duration
	ARPTimeout        time.Duration
	ARPRequestDelay   time.Duration
	AllowedPorts      []int
	MaxHostsPerSubnet int
}

func Load() Config {
	return Config{
		Addr:              getEnv("APP_ADDR", ":8080"),
		DBPath:            getEnv("APP_DB_PATH", "data/lan-topology-mapper.db"),
		StaticDir:         getEnv("APP_STATIC_DIR", "web/dist"),
		SubnetOverride:    strings.TrimSpace(os.Getenv("APP_SUBNET")),
		Concurrency:       getEnvInt("APP_SCAN_CONCURRENCY", 64),
		ConnectTimeout:    time.Duration(getEnvInt("APP_CONNECT_TIMEOUT_MS", 400)) * time.Millisecond,
		ARPTimeout:        time.Duration(getEnvInt("APP_ARP_TIMEOUT_MS", 1500)) * time.Millisecond,
		ARPRequestDelay:   time.Duration(getEnvInt("APP_ARP_REQUEST_DELAY_MS", 5)) * time.Millisecond,
		AllowedPorts:      getEnvPorts("APP_SCAN_PORTS", []int{22, 53, 80, 443, 445, 8080, 8443, 3389}),
		MaxHostsPerSubnet: getEnvInt("APP_MAX_HOSTS_PER_SUBNET", 256),
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvPorts(key string, fallback []int) []int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	ports := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		port, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		ports = append(ports, port)
	}

	if len(ports) == 0 {
		return fallback
	}
	return ports
}
