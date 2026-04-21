# LAN Topology Mapper

Linux-only LAN discovery tool with active raw ARP as the primary subnet discovery method.

## Included

- Go API server
- SQLite persistence
- Linux subnet and default gateway detection
- Active raw ARP sweep across the local subnet
- ICMP and bounded TCP connect host discovery for secondary evidence and service detection
- Neighbor-cache (`/proc/net/arp`) reconciliation
- Reverse DNS hostname lookup
- Local OUI vendor lookup seed data
- React dashboard with scan summary and device table
- Dockerized build and runtime flow

## Notes

- Best discovery requires Linux raw socket access.
- Primary LAN presence signal is active ARP.
- ICMP and TCP are used as secondary evidence and enrichment.

## Run With Docker

```bash
docker build -t lan-topology-mapper .
docker run --rm --network host --cap-add NET_RAW -v "$(pwd)/data:/app/data" lan-topology-mapper
```

Then open `http://localhost:8080`.

Best discovery mode runs the container as root with `NET_RAW`, which is the default in the image.

## Useful Environment Variables

- `APP_ADDR`
- `APP_DB_PATH`
- `APP_STATIC_DIR`
- `APP_SUBNET` for manual subnet override, for example `192.168.1.0/24`
- `APP_SCAN_CONCURRENCY`
- `APP_CONNECT_TIMEOUT_MS`
- `APP_ARP_TIMEOUT_MS`
- `APP_ARP_REQUEST_DELAY_MS`
- `APP_SCAN_PORTS`
- `APP_MAX_HOSTS_PER_SUBNET`
