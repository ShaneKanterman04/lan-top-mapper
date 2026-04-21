package discovery

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/mdlayher/arp"

	"lan-topology-mapper/internal/config"
	"lan-topology-mapper/internal/models"
)

type Scanner struct {
	config config.Config

	serviceMap map[int]string
	vendors    map[string]string
	pingPath   string
	ipPath     string
}

type Plan struct {
	Subnet      string
	GatewayIP   string
	Targets     []models.ScanTarget
	Interface   *net.Interface
	Mode        string
	LimitedMode bool
	ActiveARP   bool
}

type Result struct {
	Observations []models.DeviceObservation
	Diagnostics  models.ScanDiagnostics
}

func NewScanner(cfg config.Config) *Scanner {
	pingPath, _ := exec.LookPath("ping")
	ipPath, _ := exec.LookPath("ip")

	return &Scanner{
		config: cfg,
		serviceMap: map[int]string{
			22:   "SSH",
			53:   "DNS",
			80:   "HTTP",
			443:  "HTTPS",
			445:  "SMB",
			8080: "HTTP-Alt",
			8443: "HTTPS-Alt",
			3389: "RDP",
		},
		vendors: map[string]string{
			"B827EB": "Raspberry Pi Foundation",
			"DCA632": "Raspberry Pi Trading",
			"18B430": "eero",
			"0C8BFD": "eero",
			"44A56E": "eero",
			"9C8E99": "eero",
			"AC63BE": "eero",
			"F4F5DB": "Google",
			"3C5A37": "Google Nest",
			"00163E": "Cisco",
			"A4B197": "Apple",
			"FCFBFB": "Apple",
			"000C29": "VMware",
			"F8E43B": "Ubiquiti",
			"3C52A1": "Ubiquiti",
			"E4956E": "Samsung",
			"7085C2": "Intel",
			"DCA904": "Intel",
		},
		pingPath: pingPath,
		ipPath:   ipPath,
	}
}

func (s *Scanner) Capabilities() models.ScanDiagnostics {
	availableSources := []string{"neighbor-cache", "tcp", "mdns", "ssdp", "upnp-desc"}
	if s.pingPath != "" {
		availableSources = append(availableSources, "icmp")
	}
	mode := "limited"
	if iface, _, _, err := detectPrimaryNetwork(s.config.SubnetOverride); err == nil && iface != nil && s.canUseActiveARP(iface) {
		availableSources = append([]string{"arp-active"}, availableSources...)
		mode = "enhanced"
	}
	return models.ScanDiagnostics{Mode: mode, AvailableSources: availableSources}
}

func (s *Scanner) PreparePlan() (Plan, error) {
	iface, ipNet, gatewayIP, err := detectPrimaryNetwork(s.config.SubnetOverride)
	if err != nil {
		return Plan{}, err
	}

	targets, err := hostsFromCIDR(ipNet, s.config.MaxHostsPerSubnet)
	if err != nil {
		return Plan{}, err
	}

	neighbors := readARPNeighbors()
	planned := make([]models.ScanTarget, 0, len(targets))
	for _, ip := range targets {
		mac := neighbors[ip.String()]
		source := "subnet"
		if mac != "" {
			source = "neighbor-cache"
		}
		planned = append(planned, models.ScanTarget{IP: ip.String(), MAC: mac, Source: source})
	}

	activeARP := iface != nil && s.canUseActiveARP(iface)
	mode := "limited"
	if activeARP {
		mode = "enhanced"
	}

	return Plan{
		Subnet:      ipNet.String(),
		GatewayIP:   gatewayIP,
		Targets:     planned,
		Interface:   iface,
		Mode:        mode,
		LimitedMode: !activeARP,
		ActiveARP:   activeARP,
	}, nil
}

func (s *Scanner) Scan(ctx context.Context, plan Plan) (Result, error) {
	startedAt := time.Now()
	activeARPResults := map[string]string{}
	if plan.ActiveARP && plan.Interface != nil {
		activeARPResults = s.runActiveARPSweep(ctx, plan.Interface, plan.Targets)
	}
	identityClaimsByIP := mergeIdentityClaimMaps(
		s.runMDNSBrowse(ctx),
		s.runSSDPDiscovery(ctx),
	)

	enrichmentTargets := s.prepareEnrichmentTargets(plan.Targets, activeARPResults, identityClaimsByIP, plan.GatewayIP)
	jobs := make(chan models.ScanTarget)
	results := make(chan models.DeviceObservation)

	workerCount := s.config.Concurrency
	if workerCount < 1 {
		workerCount = 1
	}

	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for target := range jobs {
				observation, ok := s.scanHost(ctx, target)
				if !ok {
					continue
				}
				select {
				case results <- observation:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, target := range enrichmentTargets {
			select {
			case jobs <- target:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	observationsByIP := map[string]models.DeviceObservation{}
	for ip, mac := range activeARPResults {
		observation := models.DeviceObservation{
			IP:             ip,
			MAC:            mac,
			Vendor:         vendorForMAC(mac, s.vendors),
			Online:         true,
			Evidence:       []string{"arp-active"},
			IdentityClaims: append([]models.IdentityClaim(nil), identityClaimsByIP[ip]...),
		}
		observation.Hostname = lookupHostname(ip)
		observationsByIP[ip] = observation
	}

	for {
		select {
		case observation, ok := <-results:
			if !ok {
				observations := mapsToSortedObservations(observationsByIP)
				availableSources := []string{"neighbor-cache", "tcp", "mdns", "ssdp", "upnp-desc"}
				if s.pingPath != "" {
					availableSources = append(availableSources, "icmp")
				}
				if plan.ActiveARP {
					availableSources = append([]string{"arp-active"}, availableSources...)
				}
				return Result{
					Observations: observations,
					Diagnostics: models.ScanDiagnostics{
						Mode:             plan.Mode,
						AvailableSources: availableSources,
						EvidenceCounts:   summarizeEvidence(observations),
						TotalTargets:     len(plan.Targets),
						DevicesFound:     len(observations),
						DurationMs:       time.Since(startedAt).Milliseconds(),
					},
				}, nil
			}
			mergeObservation(observationsByIP, observation)
		case <-ctx.Done():
			return Result{}, ctx.Err()
		}
	}
}

func (s *Scanner) prepareEnrichmentTargets(base []models.ScanTarget, activeARPResults map[string]string, identityClaimsByIP map[string][]models.IdentityClaim, gatewayIP string) []models.ScanTarget {
	if len(activeARPResults) == 0 {
		for i := range base {
			base[i].Claims = append(base[i].Claims, identityClaimsByIP[base[i].IP]...)
		}
		return base
	}

	targetMap := map[string]models.ScanTarget{}
	for _, target := range base {
		target.Claims = append(target.Claims, identityClaimsByIP[target.IP]...)
		if mac, ok := activeARPResults[target.IP]; ok {
			target.MAC = mac
			target.Source = "arp-active"
			targetMap[target.IP] = target
			continue
		}
		if target.Source == "neighbor-cache" || target.IP == gatewayIP {
			targetMap[target.IP] = target
		}
	}

	for ip, mac := range activeARPResults {
		if _, ok := targetMap[ip]; !ok {
			targetMap[ip] = models.ScanTarget{IP: ip, MAC: mac, Source: "arp-active", Claims: append([]models.IdentityClaim(nil), identityClaimsByIP[ip]...)}
		}
	}

	targets := make([]models.ScanTarget, 0, len(targetMap))
	for _, target := range targetMap {
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].IP < targets[j].IP })
	return targets
}

func (s *Scanner) scanHost(ctx context.Context, target models.ScanTarget) (models.DeviceObservation, bool) {
	evidence := map[string]struct{}{}
	mac := target.MAC
	switch target.Source {
	case "arp-active":
		evidence["arp-active"] = struct{}{}
	case "neighbor-cache":
		evidence["neighbor-cache"] = struct{}{}
	}

	if s.pingPath != "" && s.tryPing(ctx, target.IP) {
		evidence["icmp"] = struct{}{}
	}

	services := make([]models.Service, 0, len(s.config.AllowedPorts))
	for _, port := range s.config.AllowedPorts {
		if s.tryConnect(ctx, target.IP, port) {
			services = append(services, models.Service{
				Port:        port,
				Protocol:    "tcp",
				ServiceName: s.serviceNameForPort(port),
			})
			evidence["tcp"] = struct{}{}
		}
	}

	if updatedMAC := s.readNeighborForIP(target.IP); updatedMAC != "" {
		if mac == "" {
			mac = updatedMAC
		}
		if _, ok := evidence["arp-active"]; !ok {
			evidence["neighbor-cache"] = struct{}{}
		}
	}

	if len(evidence) == 0 {
		return models.DeviceObservation{}, false
	}

	hostname := lookupHostname(target.IP)
	vendor := vendorForMAC(mac, s.vendors)
	claims := append([]models.IdentityClaim{}, target.Claims...)
	if hostname != "" {
		claims = append(claims, models.IdentityClaim{Source: "reverse-dns", ClaimType: "hostname", Value: hostname, ConfidenceLabel: "low"})
	}
	if vendor != "" {
		claims = append(claims, models.IdentityClaim{Source: "oui", ClaimType: "manufacturer", Value: vendor, ConfidenceLabel: "low"})
	}
	claims = append(claims, s.probeHTTPIdentity(target.IP, services)...)
	claims = dedupeIdentityClaims(claims)
	displayName, nameSource, resolvedVendor, vendorSource, deviceType, typeSource, confidence, reasons := s.classifyDevice(target.IP, hostname, vendor, services, claims)
	if resolvedVendor != "" {
		vendor = resolvedVendor
	}

	return models.DeviceObservation{
		IP:                       target.IP,
		MAC:                      mac,
		Hostname:                 hostname,
		Vendor:                   vendor,
		DisplayName:              displayName,
		NameSource:               nameSource,
		DeviceType:               deviceType,
		TypeSource:               typeSource,
		ClassificationConfidence: confidence,
		ClassificationReasons:    reasons,
		VendorSource:             vendorSource,
		IdentityClaims:           claims,
		Online:                   true,
		Services:                 services,
		Evidence:                 sortedEvidence(evidence),
	}, true
}

func (s *Scanner) classifyDevice(ip, hostname, vendor string, services []models.Service, claims []models.IdentityClaim) (string, string, string, string, string, string, string, []string) {
	reasons := make([]string, 0)
	servicePorts := map[int]struct{}{}
	for _, service := range services {
		servicePorts[service.Port] = struct{}{}
	}

	identityTextParts := []string{hostname, vendor}
	bestName, nameSource := bestClaim(claims, map[string]int{
		"mdns:display_name":      100,
		"upnp-desc:display_name": 95,
		"mdns:hostname":          90,
		"nbns:hostname":          88,
		"upnp-desc:model":        80,
		"http:http_title":        70,
		"tls:tls_name":           65,
		"reverse-dns:hostname":   60,
	})
	bestVendor, vendorSource := bestClaim(claims, map[string]int{
		"upnp-desc:manufacturer": 90,
		"mdns:manufacturer":      80,
		"oui:manufacturer":       40,
	})
	if bestVendor == "" {
		bestVendor, vendorSource = vendor, firstNonEmpty(vendorSource, "oui")
	}
	if bestName == "" {
		bestName = firstNonEmpty(hostname, bestVendor, ip)
		nameSource = fallbackSource(bestName, hostname, bestVendor)
	}
	for _, claim := range claims {
		identityTextParts = append(identityTextParts, claim.Value)
		if claim.ClaimType == "display_name" || claim.ClaimType == "manufacturer" || claim.ClaimType == "model" || claim.ClaimType == "hostname" || claim.ClaimType == "server" || claim.ClaimType == "device_type" {
			reasons = append(reasons, fmt.Sprintf("%s=%s", claim.Source, claim.Value))
		}
	}

	displayName := bestName
	deviceType := "Unknown device"
	typeSource := "heuristic"
	confidence := "low"
	identityText := strings.ToLower(strings.Join(identityTextParts, " "))

	switch {
	case strings.Contains(identityText, "eero"):
		deviceType = "Mesh router"
		typeSource = "claim"
		confidence = "high"
	case strings.Contains(identityText, "printer") || hasAnyPort(servicePorts, 631, 9100):
		deviceType = "Printer"
		typeSource = "claim"
		confidence = "high"
	case strings.Contains(identityText, "chromecast") || strings.Contains(identityText, "google tv"):
		deviceType = "Media device"
		typeSource = "claim"
		confidence = "high"
	case strings.Contains(identityText, "nas") || hasAnyPort(servicePorts, 445):
		deviceType = "NAS or file server"
		confidence = "medium"
	case hasAnyPort(servicePorts, 53) && hasAnyPort(servicePorts, 80, 443, 8443):
		deviceType = "Router or access point"
		confidence = "medium"
	case hasAnyPort(servicePorts, 22) && len(servicePorts) <= 2:
		deviceType = "Server or appliance"
		confidence = "medium"
	case strings.Contains(identityText, "apple") && strings.Contains(identityText, "tv"):
		deviceType = "Streaming device"
		confidence = "medium"
	case strings.Contains(identityText, "apple"):
		deviceType = "Apple device"
		confidence = "medium"
	case strings.Contains(identityText, "intel") || strings.Contains(identityText, "windows"):
		deviceType = "Computer"
		confidence = "low"
	}

	if displayName == bestVendor && deviceType == "Unknown device" {
		confidence = "low"
	}
	if typeClaim, source := bestClaim(claims, map[string]int{
		"upnp-desc:device_type": 95,
		"upnp-desc:model":       85,
		"mdns:device_type":      80,
	}); typeClaim != "" && deviceType == "Unknown device" {
		deviceType = typeClaim
		typeSource = source
		confidence = "medium"
	}
	return displayName, nameSource, bestVendor, vendorSource, deviceType, typeSource, confidence, dedupeStrings(reasons)
}

func (s *Scanner) probeHTTPIdentity(ip string, services []models.Service) []models.IdentityClaim {
	ports := []int{}
	for _, service := range services {
		switch service.Port {
		case 80, 443, 8080, 8443:
			ports = append(ports, service.Port)
		}
	}
	if len(ports) == 0 {
		return nil
	}

	client := &http.Client{
		Timeout: 1200 * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
	}

	var claims []models.IdentityClaim
	for _, port := range ports {
		scheme := "http"
		if port == 443 || port == 8443 {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(ip, strconv.Itoa(port)))
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		request.Header.Set("User-Agent", "lan-topology-mapper/1.0")
		response, err := client.Do(request)
		if err != nil {
			continue
		}
		server := strings.TrimSpace(response.Header.Get("Server"))
		if server != "" {
			claims = append(claims, models.IdentityClaim{Source: "http", ClaimType: "server", Value: server, ConfidenceLabel: "low"})
		}
		if response.TLS != nil && len(response.TLS.PeerCertificates) > 0 {
			certName := strings.TrimSpace(response.TLS.PeerCertificates[0].Subject.CommonName)
			if certName == "" && len(response.TLS.PeerCertificates[0].DNSNames) > 0 {
				certName = strings.TrimSpace(response.TLS.PeerCertificates[0].DNSNames[0])
			}
			if certName != "" {
				claims = append(claims, models.IdentityClaim{Source: "tls", ClaimType: "tls_name", Value: certName, ConfidenceLabel: "low"})
			}
		}
		{
			body := make([]byte, 4096)
			count, _ := response.Body.Read(body)
			title := extractHTMLTitle(string(body[:count]))
			if title != "" {
				claims = append(claims, models.IdentityClaim{Source: "http", ClaimType: "display_name", Value: title, ConfidenceLabel: "medium"})
				claims = append(claims, models.IdentityClaim{Source: "http", ClaimType: "http_title", Value: title, ConfidenceLabel: "medium"})
			}
		}
		_ = response.Body.Close()
		if len(claims) > 0 {
			break
		}
	}
	return claims
}

func (s *Scanner) runMDNSBrowse(ctx context.Context) map[string][]models.IdentityClaim {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil
	}
	serviceTypes := []string{"_workstation._tcp", "_http._tcp", "_ipp._tcp", "_airplay._tcp", "_googlecast._tcp", "_smb._tcp"}
	claimsByIP := map[string][]models.IdentityClaim{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, serviceType := range serviceTypes {
		entries := make(chan *zeroconf.ServiceEntry)
		browseCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		wg.Add(1)
		go func(st string, browseCtx context.Context, ch <-chan *zeroconf.ServiceEntry) {
			defer wg.Done()
			for {
				select {
				case <-browseCtx.Done():
					return
				case entry := <-ch:
					if entry == nil {
						continue
					}
					for _, ip := range entry.AddrIPv4 {
						key := ip.String()
						claims := []models.IdentityClaim{}
						if entry.Instance != "" {
							claims = append(claims, models.IdentityClaim{Source: "mdns", ClaimType: "display_name", Value: entry.Instance, ConfidenceLabel: "high"})
						}
						if entry.HostName != "" {
							claims = append(claims, models.IdentityClaim{Source: "mdns", ClaimType: "hostname", Value: strings.TrimSuffix(entry.HostName, "."), ConfidenceLabel: "high"})
						}
						claims = append(claims, models.IdentityClaim{Source: "mdns", ClaimType: "service", Value: st, ConfidenceLabel: "medium"})
						for _, text := range entry.Text {
							lower := strings.ToLower(text)
							switch {
							case strings.HasPrefix(lower, "model="):
								claims = append(claims, models.IdentityClaim{Source: "mdns", ClaimType: "model", Value: textValue(text), ConfidenceLabel: "medium"})
							case strings.HasPrefix(lower, "ty="):
								claims = append(claims, models.IdentityClaim{Source: "mdns", ClaimType: "device_type", Value: textValue(text), ConfidenceLabel: "medium"})
							case strings.HasPrefix(lower, "manufacturer="):
								claims = append(claims, models.IdentityClaim{Source: "mdns", ClaimType: "manufacturer", Value: textValue(text), ConfidenceLabel: "medium"})
							}
						}
						mu.Lock()
						claimsByIP[key] = append(claimsByIP[key], claims...)
						mu.Unlock()
					}
				}
			}
		}(serviceType, browseCtx, entries)
		go func(st string, c context.Context, cancelFn context.CancelFunc) {
			defer cancelFn()
			_ = resolver.Browse(c, st, "local.", entries)
		}(serviceType, browseCtx, cancel)
	}

	wg.Wait()
	for ip := range claimsByIP {
		claimsByIP[ip] = dedupeIdentityClaims(claimsByIP[ip])
	}
	return claimsByIP
}

func (s *Scanner) runSSDPDiscovery(ctx context.Context) map[string][]models.IdentityClaim {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil
	}
	defer conn.Close()
	message := strings.Join([]string{
		"M-SEARCH * HTTP/1.1",
		"HOST: 239.255.255.250:1900",
		"MAN: \"ssdp:discover\"",
		"MX: 1",
		"ST: ssdp:all",
		"",
		"",
	}, "\r\n")
	_, _ = conn.WriteTo([]byte(message), &net.UDPAddr{IP: net.ParseIP("239.255.255.250"), Port: 1900})

	claimsByIP := map[string][]models.IdentityClaim{}
	buffer := make([]byte, 8192)
	deadline := time.Now().Add(1500 * time.Millisecond)
	_ = conn.SetReadDeadline(deadline)
	for {
		if ctx.Err() != nil {
			break
		}
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			if isDeadline(err) {
				break
			}
			break
		}
		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok || udpAddr.IP == nil {
			continue
		}
		ip := udpAddr.IP.String()
		headers := parseSSDPHeaders(string(buffer[:n]))
		claims := []models.IdentityClaim{}
		if server := headers["server"]; server != "" {
			claims = append(claims, models.IdentityClaim{Source: "ssdp", ClaimType: "server", Value: server, ConfidenceLabel: "low"})
		}
		if st := headers["st"]; st != "" {
			claims = append(claims, models.IdentityClaim{Source: "ssdp", ClaimType: "device_type", Value: st, ConfidenceLabel: "low"})
		}
		if usn := headers["usn"]; usn != "" {
			claims = append(claims, models.IdentityClaim{Source: "ssdp", ClaimType: "usn", Value: usn, ConfidenceLabel: "low"})
		}
		if location := headers["location"]; location != "" {
			claims = append(claims, models.IdentityClaim{Source: "ssdp", ClaimType: "location", Value: location, ConfidenceLabel: "low"})
			claims = append(claims, s.fetchUPnPDescription(location)...)
		}
		if len(claims) > 0 {
			claimsByIP[ip] = append(claimsByIP[ip], claims...)
		}
	}
	for ip := range claimsByIP {
		claimsByIP[ip] = dedupeIdentityClaims(claimsByIP[ip])
	}
	return claimsByIP
}

func (s *Scanner) fetchUPnPDescription(location string) []models.IdentityClaim {
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	response, err := client.Get(location)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 65536))
	if err != nil {
		return nil
	}
	var root upnpRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil
	}
	claims := make([]models.IdentityClaim, 0)
	if root.Device.FriendlyName != "" {
		claims = append(claims, models.IdentityClaim{Source: "upnp-desc", ClaimType: "display_name", Value: root.Device.FriendlyName, ConfidenceLabel: "high", RawJSON: location})
	}
	if root.Device.Manufacturer != "" {
		claims = append(claims, models.IdentityClaim{Source: "upnp-desc", ClaimType: "manufacturer", Value: root.Device.Manufacturer, ConfidenceLabel: "high", RawJSON: location})
	}
	if root.Device.ModelName != "" {
		claims = append(claims, models.IdentityClaim{Source: "upnp-desc", ClaimType: "model", Value: root.Device.ModelName, ConfidenceLabel: "high", RawJSON: location})
	}
	if root.Device.DeviceType != "" {
		claims = append(claims, models.IdentityClaim{Source: "upnp-desc", ClaimType: "device_type", Value: root.Device.DeviceType, ConfidenceLabel: "medium", RawJSON: location})
	}
	return claims
}

func (s *Scanner) runActiveARPSweep(ctx context.Context, iface *net.Interface, targets []models.ScanTarget) map[string]string {
	client, err := arp.Dial(iface)
	if err != nil {
		return map[string]string{}
	}
	defer client.Close()

	results := map[string]string{}
	var mu sync.Mutex
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			_ = client.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
			packet, _, err := client.Read()
			if err != nil {
				if isDeadline(err) {
					continue
				}
				return
			}
			if packet.Operation != arp.OperationReply {
				continue
			}
			mu.Lock()
			results[packet.SenderIP.String()] = strings.ToUpper(packet.SenderHardwareAddr.String())
			mu.Unlock()
		}
	}()

	for _, target := range targets {
		select {
		case <-ctx.Done():
			close(done)
			return results
		default:
		}

		addr, err := netip.ParseAddr(target.IP)
		if err != nil || !addr.Is4() {
			continue
		}
		_ = client.Request(addr)
		if s.config.ARPRequestDelay > 0 {
			timer := time.NewTimer(s.config.ARPRequestDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				close(done)
				return results
			case <-timer.C:
			}
		}
	}

	deadline := time.NewTimer(s.config.ARPTimeout)
	select {
	case <-ctx.Done():
	case <-deadline.C:
	}
	deadline.Stop()
	close(done)
	return results
}

func (s *Scanner) canUseActiveARP(iface *net.Interface) bool {
	if iface == nil {
		return false
	}
	client, err := arp.Dial(iface)
	if err != nil {
		return false
	}
	_ = client.Close()
	return true
}

func (s *Scanner) tryPing(ctx context.Context, ip string) bool {
	if s.pingPath == "" {
		return false
	}
	command := exec.CommandContext(ctx, s.pingPath, "-n", "-c", "1", "-W", "1", ip)
	if err := command.Run(); err != nil {
		return false
	}
	return true
}

func (s *Scanner) tryConnect(ctx context.Context, ip string, port int) bool {
	dialer := &net.Dialer{Timeout: s.config.ConnectTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (s *Scanner) readNeighborForIP(ip string) string {
	if mac := readARPNeighbors()[ip]; mac != "" {
		return mac
	}
	if s.ipPath == "" {
		return ""
	}
	output, err := exec.Command(s.ipPath, "neigh", "show", ip).Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(output))
	for i := 0; i < len(fields)-1; i++ {
		if fields[i] == "lladdr" {
			return strings.ToUpper(fields[i+1])
		}
	}
	return ""
}

func (s *Scanner) serviceNameForPort(port int) string {
	if name, ok := s.serviceMap[port]; ok {
		return name
	}
	return fmt.Sprintf("TCP/%d", port)
}

func summarizeEvidence(observations []models.DeviceObservation) map[string]int {
	counts := map[string]int{}
	for _, observation := range observations {
		for _, source := range observation.Evidence {
			counts[source]++
		}
	}
	return counts
}

func sortedEvidence(evidence map[string]struct{}) []string {
	items := make([]string, 0, len(evidence))
	for source := range evidence {
		items = append(items, source)
	}
	sort.Strings(items)
	return items
}

func mergeObservation(observations map[string]models.DeviceObservation, next models.DeviceObservation) {
	current, ok := observations[next.IP]
	if !ok {
		observations[next.IP] = next
		return
	}

	if current.MAC == "" && next.MAC != "" {
		current.MAC = next.MAC
	}
	if current.Hostname == "" && next.Hostname != "" {
		current.Hostname = next.Hostname
	}
	if current.Vendor == "" && next.Vendor != "" {
		current.Vendor = next.Vendor
	}
	if current.DisplayName == "" && next.DisplayName != "" {
		current.DisplayName = next.DisplayName
	}
	if current.NameSource == "" && next.NameSource != "" {
		current.NameSource = next.NameSource
	}
	if current.DeviceType == "" && next.DeviceType != "" {
		current.DeviceType = next.DeviceType
	}
	if current.TypeSource == "" && next.TypeSource != "" {
		current.TypeSource = next.TypeSource
	}
	if current.ClassificationConfidence == "" && next.ClassificationConfidence != "" {
		current.ClassificationConfidence = next.ClassificationConfidence
	}
	if len(next.ClassificationReasons) > 0 {
		current.ClassificationReasons = mergeEvidence(current.ClassificationReasons, next.ClassificationReasons)
	}
	if current.VendorSource == "" && next.VendorSource != "" {
		current.VendorSource = next.VendorSource
	}
	if len(next.IdentityClaims) > 0 {
		current.IdentityClaims = dedupeIdentityClaims(append(current.IdentityClaims, next.IdentityClaims...))
	}
	if len(next.Services) > 0 {
		current.Services = mergeServices(current.Services, next.Services)
	}
	current.Online = current.Online || next.Online
	current.Evidence = mergeEvidence(current.Evidence, next.Evidence)
	observations[next.IP] = current
}

func mergeEvidence(current, next []string) []string {
	set := map[string]struct{}{}
	for _, value := range current {
		set[value] = struct{}{}
	}
	for _, value := range next {
		set[value] = struct{}{}
	}
	return sortedEvidence(set)
}

func mergeServices(current, next []models.Service) []models.Service {
	seen := map[string]models.Service{}
	for _, service := range current {
		seen[fmt.Sprintf("%s:%d", service.Protocol, service.Port)] = service
	}
	for _, service := range next {
		seen[fmt.Sprintf("%s:%d", service.Protocol, service.Port)] = service
	}
	services := make([]models.Service, 0, len(seen))
	for _, service := range seen {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Port < services[j].Port })
	return services
}

func mapsToSortedObservations(values map[string]models.DeviceObservation) []models.DeviceObservation {
	observations := make([]models.DeviceObservation, 0, len(values))
	for _, observation := range values {
		observations = append(observations, observation)
	}
	sort.Slice(observations, func(i, j int) bool { return observations[i].IP < observations[j].IP })
	return observations
}

func detectPrimaryNetwork(override string) (*net.Interface, *net.IPNet, string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, "", fmt.Errorf("list interfaces: %w", err)
	}

	if override != "" {
		ip, ipNet, err := net.ParseCIDR(override)
		if err != nil {
			return nil, nil, "", fmt.Errorf("parse subnet override: %w", err)
		}
		ipNet.IP = ip.Mask(ipNet.Mask)
		iface, err := interfaceForNetwork(interfaces, ipNet)
		if err != nil {
			return nil, nil, "", err
		}
		return iface, ipNet, gatewayFromRouteFile(), nil
	}

	gateway := gatewayFromRouteFile()
	if gateway == "" {
		return nil, nil, "", errors.New("unable to detect default gateway from /proc/net/route")
	}

	gatewayIP := net.ParseIP(gateway)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			if ipNet.Contains(gatewayIP) {
				ipNet.IP = ipNet.IP.Mask(ipNet.Mask)
				ifaceCopy := iface
				return &ifaceCopy, ipNet, gateway, nil
			}
		}
	}

	return nil, nil, "", errors.New("unable to match gateway to an active IPv4 interface")
}

func interfaceForNetwork(interfaces []net.Interface, targetNet *net.IPNet) (*net.Interface, error) {
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			if targetNet.Contains(ipNet.IP) {
				ifaceCopy := iface
				return &ifaceCopy, nil
			}
		}
	}
	return nil, fmt.Errorf("unable to match subnet %s to an active IPv4 interface", targetNet.String())
}

func gatewayFromRouteFile() string {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return ""
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}
		gatewayHex := fields[2]
		if len(gatewayHex) != 8 {
			continue
		}
		parts := []string{}
		for i := 0; i < 8; i += 2 {
			value, err := strconv.ParseUint(gatewayHex[i:i+2], 16, 8)
			if err != nil {
				parts = nil
				break
			}
			parts = append([]string{strconv.Itoa(int(value))}, parts...)
		}
		if len(parts) == 4 {
			return strings.Join(parts, ".")
		}
	}

	return ""
}

func hostsFromCIDR(ipNet *net.IPNet, maxHosts int) ([]net.IP, error) {
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		return nil, errors.New("only IPv4 subnets are supported")
	}
	hostBits := bits - ones
	if hostBits <= 1 {
		return nil, errors.New("subnet has no usable host addresses")
	}
	hostCount := (1 << hostBits) - 2
	if hostCount > maxHosts {
		return nil, fmt.Errorf("subnet %s exceeds max host limit (%d > %d)", ipNet.String(), hostCount, maxHosts)
	}

	base := ipNet.IP.To4()
	if base == nil {
		return nil, errors.New("only IPv4 subnets are supported")
	}

	result := make([]net.IP, 0, hostCount)
	for i := 1; i <= hostCount; i++ {
		ip := make(net.IP, len(base))
		copy(ip, base)
		incrementIPv4(ip, i)
		result = append(result, ip)
	}
	return result, nil
}

func incrementIPv4(ip net.IP, increment int) {
	value := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	value += uint32(increment)
	ip[0] = byte(value >> 24)
	ip[1] = byte(value >> 16)
	ip[2] = byte(value >> 8)
	ip[3] = byte(value)
}

func readARPNeighbors() map[string]string {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return map[string]string{}
	}
	defer file.Close()

	neighbors := map[string]string{}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return neighbors
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		ip := fields[0]
		mac := strings.ToUpper(fields[3])
		if mac == "00:00:00:00:00:00" {
			continue
		}
		neighbors[ip] = mac
	}

	return neighbors
}

func lookupHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func vendorForMAC(mac string, vendorDB map[string]string) string {
	if mac == "" {
		return ""
	}
	normalized := strings.ReplaceAll(strings.ToUpper(mac), ":", "")
	if len(normalized) < 6 {
		return ""
	}
	return vendorDB[normalized[:6]]
}

func isDeadline(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func extractHTMLTitle(body string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return ""
	}
	value := strings.TrimSpace(strings.ReplaceAll(matches[1], "\n", " "))
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 120 {
		return value[:120]
	}
	return value
}

func hasAnyPort(ports map[int]struct{}, values ...int) bool {
	for _, value := range values {
		if _, ok := ports[value]; ok {
			return true
		}
	}
	return false
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := set[value]; ok {
			continue
		}
		set[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func textValue(value string) string {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return value
	}
	return strings.TrimSpace(parts[1])
}

func bestClaim(claims []models.IdentityClaim, weights map[string]int) (string, string) {
	bestWeight := -1
	bestValue := ""
	bestSource := ""
	for _, claim := range claims {
		key := fmt.Sprintf("%s:%s", claim.Source, claim.ClaimType)
		weight, ok := weights[key]
		if !ok || strings.TrimSpace(claim.Value) == "" {
			continue
		}
		if weight > bestWeight {
			bestWeight = weight
			bestValue = claim.Value
			bestSource = claim.Source
		}
	}
	return bestValue, bestSource
}

func fallbackSource(bestName, hostname, vendor string) string {
	switch {
	case bestName == hostname && hostname != "":
		return "reverse-dns"
	case bestName == vendor && vendor != "":
		return "oui"
	default:
		return "ip"
	}
}

func dedupeIdentityClaims(claims []models.IdentityClaim) []models.IdentityClaim {
	if len(claims) == 0 {
		return nil
	}
	seen := map[string]models.IdentityClaim{}
	for _, claim := range claims {
		key := fmt.Sprintf("%s|%s|%s", claim.Source, claim.ClaimType, claim.Value)
		seen[key] = claim
	}
	result := make([]models.IdentityClaim, 0, len(seen))
	for _, claim := range seen {
		result = append(result, claim)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Source == result[j].Source {
			if result[i].ClaimType == result[j].ClaimType {
				return result[i].Value < result[j].Value
			}
			return result[i].ClaimType < result[j].ClaimType
		}
		return result[i].Source < result[j].Source
	})
	return result
}

func mergeIdentityClaimMaps(maps ...map[string][]models.IdentityClaim) map[string][]models.IdentityClaim {
	merged := map[string][]models.IdentityClaim{}
	for _, claimMap := range maps {
		for ip, claims := range claimMap {
			merged[ip] = append(merged[ip], claims...)
		}
	}
	for ip := range merged {
		merged[ip] = dedupeIdentityClaims(merged[ip])
	}
	return merged
}

func parseSSDPHeaders(raw string) map[string]string {
	lines := strings.Split(raw, "\r\n")
	headers := map[string]string{}
	for _, line := range lines[1:] {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		headers[strings.ToLower(strings.TrimSpace(parts[0]))] = strings.TrimSpace(parts[1])
	}
	return headers
}

type upnpRoot struct {
	Device upnpDevice `xml:"device"`
}

type upnpDevice struct {
	FriendlyName string `xml:"friendlyName"`
	Manufacturer string `xml:"manufacturer"`
	ModelName    string `xml:"modelName"`
	DeviceType   string `xml:"deviceType"`
}
