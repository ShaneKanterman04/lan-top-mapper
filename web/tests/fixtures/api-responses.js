export const summaryFixture = {
  subnet: '192.168.1.0/24',
  deviceCount: 5,
  onlineCount: 4,
  currentMode: 'enhanced',
  limitedMode: false,
  gatewayIp: '192.168.1.1',
  lastScanAt: new Date(Date.now() - 3600 * 1000).toISOString(),
  latestScanId: 42,
  latestScanState: 'idle',
  latestScanDiagnostics: {
    totalTargets: 254,
    devicesFound: 5,
    durationMs: 12500,
    evidenceCounts: { arp: 3, icmp: 2, tcp: 1 },
  },
  availableSources: ['arp', 'icmp', 'tcp', 'mdns'],
}

export const emptySummaryFixture = {
  subnet: null,
  deviceCount: 0,
  onlineCount: 0,
  currentMode: 'limited',
  limitedMode: true,
  gatewayIp: null,
  lastScanAt: null,
  latestScanId: null,
  latestScanState: 'idle',
  latestScanDiagnostics: null,
  availableSources: [],
}

export const runningScanSummaryFixture = {
  ...summaryFixture,
  latestScanState: 'running',
  latestScanDiagnostics: {
    totalTargets: 254,
    devicesFound: 2,
    durationMs: 3000,
    evidenceCounts: { arp: 1, icmp: 1 },
  },
}

export const devicesFixture = [
  {
    id: 1,
    ip: '192.168.1.1',
    displayName: 'Gateway Router',
    hostname: 'router.local',
    nameSource: 'dns',
    deviceType: 'Router or access point',
    typeSource: 'macoui',
    vendor: 'Netgear',
    vendorSource: 'macoui',
    mac: 'AA:BB:CC:11:22:33',
    online: true,
    firstSeen: new Date(Date.now() - 7 * 24 * 3600 * 1000).toISOString(),
    lastSeen: new Date().toISOString(),
    evidence: ['arp', 'icmp'],
    services: [{ port: 80, protocol: 'tcp', serviceName: 'HTTP' }],
    classificationConfidence: 'high',
    classificationReasons: ['macoui match', 'gateway position'],
  },
  {
    id: 2,
    ip: '192.168.1.42',
    displayName: null,
    hostname: null,
    nameSource: 'ip',
    deviceType: 'Computer',
    typeSource: 'ports',
    vendor: 'Dell',
    vendorSource: 'macoui',
    mac: 'DD:EE:FF:44:55:66',
    online: true,
    firstSeen: new Date(Date.now() - 3 * 24 * 3600 * 1000).toISOString(),
    lastSeen: new Date().toISOString(),
    evidence: ['arp', 'tcp'],
    services: [{ port: 22, protocol: 'tcp', serviceName: 'SSH' }, { port: 445, protocol: 'tcp', serviceName: 'SMB' }],
    classificationConfidence: 'medium',
    classificationReasons: ['service fingerprint', 'macoui vendor'],
  },
  {
    id: 3,
    ip: '192.168.1.100',
    displayName: 'Living Room TV',
    hostname: 'tv.local',
    nameSource: 'mdns',
    deviceType: 'Media device',
    typeSource: 'mdns',
    vendor: 'Samsung',
    vendorSource: 'macoui',
    mac: '11:22:33:AA:BB:CC',
    online: false,
    firstSeen: new Date(Date.now() - 14 * 24 * 3600 * 1000).toISOString(),
    lastSeen: new Date(Date.now() - 2 * 24 * 3600 * 1000).toISOString(),
    evidence: ['mdns'],
    services: [{ port: 8009, protocol: 'tcp', serviceName: 'Chromecast' }],
    classificationConfidence: 'high',
    classificationReasons: ['mdns service name', 'macoui vendor'],
  },
  {
    id: 4,
    ip: '192.168.1.55',
    displayName: null,
    hostname: null,
    nameSource: 'ip',
    deviceType: null,
    typeSource: null,
    vendor: null,
    vendorSource: null,
    mac: null,
    online: true,
    firstSeen: new Date(Date.now() - 1 * 24 * 3600 * 1000).toISOString(),
    lastSeen: new Date().toISOString(),
    evidence: ['icmp'],
    services: [],
    classificationConfidence: 'low',
    classificationReasons: ['insufficient evidence'],
  },
  {
    id: 5,
    ip: '192.168.1.10',
    displayName: 'Printer',
    hostname: 'printer.local',
    nameSource: 'upnp-desc',
    deviceType: 'Printer',
    typeSource: 'upnp',
    vendor: 'HP',
    vendorSource: 'upnp',
    mac: '44:55:66:DD:EE:FF',
    online: true,
    firstSeen: new Date(Date.now() - 30 * 24 * 3600 * 1000).toISOString(),
    lastSeen: new Date().toISOString(),
    evidence: ['upnp', 'tcp'],
    services: [{ port: 9100, protocol: 'tcp', serviceName: 'RAW' }],
    classificationConfidence: 'high',
    classificationReasons: ['upnp device type', 'service port match'],
  },
]

export const emptyDevicesFixture = []

export const topologyFixture = {
  nodes: [
    {
      data: {
        id: 'gateway',
        deviceId: 1,
        kind: 'gateway',
        label: 'Gateway Router',
        displayName: 'Gateway Router',
        ip: '192.168.1.1',
        hostname: 'router.local',
        vendor: 'Netgear',
        deviceType: 'Router or access point',
        nameSource: 'dns',
        online: true,
        evidence: ['arp', 'icmp'],
        services: [{ port: 80, protocol: 'tcp', serviceName: 'HTTP' }],
        classificationConfidence: 'high',
      },
    },
    {
      data: {
        id: 'node-2',
        deviceId: 2,
        kind: 'device',
        label: '192.168.1.42',
        displayName: null,
        ip: '192.168.1.42',
        hostname: null,
        vendor: 'Dell',
        deviceType: 'Computer',
        nameSource: 'ip',
        online: true,
        evidence: ['arp', 'tcp'],
        services: [{ port: 22, protocol: 'tcp', serviceName: 'SSH' }],
        classificationConfidence: 'medium',
      },
    },
    {
      data: {
        id: 'node-3',
        deviceId: 3,
        kind: 'device',
        label: 'Living Room TV',
        displayName: 'Living Room TV',
        ip: '192.168.1.100',
        hostname: 'tv.local',
        vendor: 'Samsung',
        deviceType: 'Media device',
        nameSource: 'mdns',
        online: false,
        evidence: ['mdns'],
        services: [{ port: 8009, protocol: 'tcp', serviceName: 'Chromecast' }],
        classificationConfidence: 'high',
      },
    },
    {
      data: {
        id: 'node-4',
        deviceId: 4,
        kind: 'device',
        label: '192.168.1.55',
        displayName: null,
        ip: '192.168.1.55',
        hostname: null,
        vendor: null,
        deviceType: null,
        nameSource: 'ip',
        online: true,
        evidence: ['icmp'],
        services: [],
        classificationConfidence: 'low',
      },
    },
    {
      data: {
        id: 'node-5',
        deviceId: 5,
        kind: 'device',
        label: 'Printer',
        displayName: 'Printer',
        ip: '192.168.1.10',
        hostname: 'printer.local',
        vendor: 'HP',
        deviceType: 'Printer',
        nameSource: 'upnp-desc',
        online: true,
        evidence: ['upnp', 'tcp'],
        services: [{ port: 9100, protocol: 'tcp', serviceName: 'RAW' }],
        classificationConfidence: 'high',
      },
    },
  ],
  edges: [
    { data: { source: 'gateway', target: 'node-2' } },
    { data: { source: 'gateway', target: 'node-3' } },
    { data: { source: 'gateway', target: 'node-4' } },
    { data: { source: 'gateway', target: 'node-5' } },
  ],
}

export const emptyTopologyFixture = { nodes: [], edges: [] }

export const deviceDetailFixture = (deviceId) => {
  const device = devicesFixture.find((d) => d.id === deviceId)
  if (!device) return null

  return {
    device: {
      ...device,
      identityClaims: [
        { source: 'dns', claimType: 'hostname', value: device.hostname || device.ip, confidenceLabel: 'high' },
        { source: 'macoui', claimType: 'vendor', value: device.vendor || 'Unknown', confidenceLabel: 'medium' },
      ],
    },
    evidenceSummary: { arp: 1, icmp: 1, tcp: 1 },
    observationHistory: [
      {
        observationId: 101,
        scanId: 42,
        scanStatus: 'completed',
        observedAt: device.lastSeen,
        online: device.online,
        displayName: device.displayName,
        hostname: device.hostname,
        ip: device.ip,
        deviceType: device.deviceType,
        classificationReasons: device.classificationReasons,
        evidence: device.evidence,
        services: device.services,
      },
    ],
    scanSummary: {
      timesSeen: 1,
      firstObservedAt: device.firstSeen,
      lastObservedAt: device.lastSeen,
    },
  }
}

export const scanStartFixture = {
  status: 'started',
  scanId: 99,
  subnet: '192.168.1.0/24',
  mode: 'enhanced',
}

export const scanStatusRunningFixture = {
  id: 99,
  status: 'running',
  diagnostics: {
    totalTargets: 254,
    devicesFound: 2,
    durationMs: 3000,
    evidenceCounts: { arp: 1, icmp: 1 },
  },
}

export const scanStatusCompletedFixture = {
  id: 99,
  status: 'completed',
  diagnostics: {
    totalTargets: 254,
    devicesFound: 5,
    durationMs: 12500,
    evidenceCounts: { arp: 3, icmp: 2, tcp: 1 },
  },
}

export const scanStatusFailedFixture = {
  id: 99,
  status: 'failed',
  diagnostics: {
    totalTargets: 254,
    devicesFound: 0,
    durationMs: 0,
    evidenceCounts: {},
  },
}
