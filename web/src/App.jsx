import React, { useCallback, useEffect, useMemo, useState } from 'react'
import DeviceDetailPanel from './DeviceDetailPanel'
import TopologyGraph from './TopologyGraph'

const pollIntervalMs = 1500

function App() {
  const [summary, setSummary] = useState(null)
  const [devices, setDevices] = useState([])
  const [topology, setTopology] = useState({ nodes: [], edges: [] })
  const [selectedView, setSelectedView] = useState('dashboard')
  const [selectedNode, setSelectedNode] = useState(null)
  const [selectedDeviceId, setSelectedDeviceId] = useState(null)
  const [selectedDeviceDetail, setSelectedDeviceDetail] = useState(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [scanStatus, setScanStatus] = useState({ loading: true })
  const [filters, setFilters] = useState({ onlineOnly: false, onlyServices: false, evidence: 'all', deviceType: 'all', confidence: 'all', nameSource: 'all', search: '' })
  const [layoutName, setLayoutName] = useState('concentric')
  const [error, setError] = useState('')

  const evidenceOptions = useMemo(() => ['all', ...(summary?.availableSources || [])], [summary?.availableSources])
  const deviceTypeOptions = useMemo(() => ['all', ...new Set(devices.map((device) => device.deviceType).filter(Boolean))], [devices])
  const confidenceOptions = ['all', 'high', 'medium', 'low']
  const nameSourceOptions = useMemo(() => ['all', ...new Set(devices.map((device) => device.nameSource || 'ip').filter(Boolean))], [devices])
  const layoutOptions = [
    { value: 'concentric', label: 'Concentric' },
    { value: 'breadthfirst', label: 'Breadth-first' },
    { value: 'radial', label: 'Radial' },
  ]

  const syncSelectedNode = useCallback((nextTopology, currentSelectedNode) => {
    if (!currentSelectedNode?.id) {
      return null
    }
    const match = (nextTopology.nodes || []).find((node) => node.data?.id === currentSelectedNode.id)
    return match ? match.data : null
  }, [])

  const loadAll = useCallback(async () => {
    setScanStatus({ loading: true })
    setError('')

    try {
      const [summaryResponse, devicesResponse, topologyResponse] = await Promise.all([
        fetch('/api/summary'),
        fetch('/api/devices'),
        fetch('/api/topology'),
      ])
      const [summaryPayload, devicesPayload, topologyPayload] = await Promise.all([
        summaryResponse.json(),
        devicesResponse.json(),
        topologyResponse.json(),
      ])

      if (!summaryResponse.ok) {
        throw new Error(summaryPayload.error || 'Failed to load summary')
      }
      if (!devicesResponse.ok) {
        throw new Error(devicesPayload.error || 'Failed to load devices')
      }
      if (!topologyResponse.ok) {
        throw new Error(topologyPayload.error || 'Failed to load topology')
      }

      setSummary(summaryPayload)
      setDevices(devicesPayload)
      setTopology(topologyPayload)
      setSelectedNode((current) => syncSelectedNode(topologyPayload, current))
    } catch (loadError) {
      setError(loadError.message)
    } finally {
      setScanStatus({ loading: false })
    }
  }, [syncSelectedNode])

  useEffect(() => {
    loadAll()
  }, [loadAll])

  useEffect(() => {
    if (!summary?.latestScanId || summary.latestScanState !== 'running') {
      return undefined
    }

    const timer = window.setInterval(async () => {
      try {
        const response = await fetch(`/api/scans/${summary.latestScanId}`)
        const payload = await response.json()
        if (!response.ok) {
          throw new Error(payload.error || 'Failed to refresh scan status')
        }

        if (payload.status !== 'running') {
          await loadAll()
        } else {
          setSummary((current) => current ? {
            ...current,
            latestScanState: payload.status,
            latestScanDiagnostics: payload.diagnostics,
          } : current)
        }
      } catch (scanError) {
        setError(scanError.message)
      }
    }, pollIntervalMs)

    return () => window.clearInterval(timer)
  }, [loadAll, summary?.latestScanId, summary?.latestScanState])

  useEffect(() => {
    if (!selectedDeviceId) {
      setSelectedDeviceDetail(null)
      setDetailError('')
      setDetailLoading(false)
      return
    }

    let cancelled = false
    setDetailLoading(true)
    setDetailError('')

    fetch(`/api/devices/${selectedDeviceId}`)
      .then(async (response) => {
        const payload = await response.json()
        if (!response.ok) {
          throw new Error(payload.error || 'Failed to load device detail')
        }
        return payload
      })
      .then((payload) => {
        if (!cancelled) {
          setSelectedDeviceDetail(payload)
        }
      })
      .catch((loadError) => {
        if (!cancelled) {
          setDetailError(loadError.message)
          setSelectedDeviceDetail(null)
        }
      })
      .finally(() => {
        if (!cancelled) {
          setDetailLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [selectedDeviceId])

  async function startScan() {
    setError('')
    setScanStatus({ loading: true })

    try {
      const response = await fetch('/api/scans', { method: 'POST' })
      const payload = await response.json()
      if (!response.ok) {
        throw new Error(payload.error || 'Failed to start scan')
      }

      setSummary((current) => ({
        ...(current || {}),
        latestScanId: payload.scanId,
        latestScanState: 'running',
        subnet: payload.subnet,
        currentMode: payload.mode,
      }))
    } catch (scanError) {
      setError(scanError.message)
    } finally {
      setScanStatus({ loading: false })
    }
  }

  function handleDeviceSelect(deviceId) {
    setSelectedDeviceId(deviceId)
  }

  function handleNodeSelect(nodeData) {
    setSelectedNode(nodeData)
    if (nodeData?.deviceId) {
      setSelectedDeviceId(nodeData.deviceId)
    }
  }

  const diagnostics = summary?.latestScanDiagnostics
  const cards = [
    { label: 'Subnet', value: summary?.subnet || 'Not scanned yet' },
    { label: 'Devices', value: summary?.deviceCount ?? 0 },
    { label: 'Online', value: summary?.onlineCount ?? 0 },
    { label: 'Mode', value: formatLabel(summary?.currentMode || 'limited') },
  ]

  return (
    <div className="app-shell">
      <header className="hero">
        <div>
          <p className="eyebrow">Linux Discovery Diagnostics</p>
          <h1>LAN Topology Mapper</h1>
          <p className="subtitle">
            Run repeatable LAN scans, inspect discovery evidence, and explore the current topology graph.
          </p>
        </div>

        <div className="hero-actions">
          <button data-testid="start-scan-button" onClick={startScan} disabled={scanStatus.loading || summary?.latestScanState === 'running'}>
            {summary?.latestScanState === 'running' ? 'Scanning...' : 'Start Scan'}
          </button>
          <span data-testid="status-pill" className="status-pill">{summary?.latestScanState || 'idle'}</span>
        </div>
      </header>

      <nav className="view-switcher">
        <button data-testid="tab-dashboard" className={selectedView === 'dashboard' ? 'tab-button tab-active' : 'tab-button'} onClick={() => setSelectedView('dashboard')}>Dashboard</button>
        <button data-testid="tab-topology" className={selectedView === 'topology' ? 'tab-button tab-active' : 'tab-button'} onClick={() => setSelectedView('topology')}>Topology</button>
      </nav>

      <section className="notice">
        {summary?.limitedMode
          ? 'Limited mode: raw ARP is unavailable, so scans fall back to ICMP, TCP probes, and neighbor-cache evidence.'
          : 'Enhanced mode: raw active ARP is the primary LAN discovery method, with ICMP, TCP, and neighbor-cache data used for enrichment.'}
      </section>

      {error && <section data-testid="error-box" className="error-box">{error}</section>}

      {selectedView === 'dashboard' ? (
        <>
          <section className="card-grid">
            {cards.map((card) => (
              <article data-testid={`metric-card-${card.label.toLowerCase().replace(/\s+/g, '-')}`} className="metric-card" key={card.label}>
                <span>{card.label}</span>
                <strong>{card.value}</strong>
              </article>
            ))}
          </section>

          <section className="panel-row">
            <article className="panel">
              <h2>Scan Snapshot</h2>
              <dl className="details-grid">
                <div>
                  <dt>Gateway</dt>
                  <dd>{summary?.gatewayIp || 'Unavailable'}</dd>
                </div>
                <div>
                  <dt>Last Scan</dt>
                  <dd>{formatTimestamp(summary?.lastScanAt)}</dd>
                </div>
                <div>
                  <dt>Available Sources</dt>
                  <dd>{summary?.availableSources?.length ? summary.availableSources.join(', ') : 'Unavailable'}</dd>
                </div>
                <div>
                  <dt>Targets</dt>
                  <dd>{diagnostics?.totalTargets ?? 0}</dd>
                </div>
              </dl>
            </article>

            <article className="panel topology-preview">
              <h2>Discovery Diagnostics</h2>
              <div className="diagnostic-stack">
                <p>Devices found: <strong>{diagnostics?.devicesFound ?? 0}</strong></p>
                <p>Duration: <strong>{formatDuration(diagnostics?.durationMs)}</strong></p>
                <div className="evidence-group">
                  {(Object.entries(diagnostics?.evidenceCounts || {})).length === 0
                    ? <span className="chip chip-offline">No evidence yet</span>
                    : Object.entries(diagnostics.evidenceCounts).map(([source, count]) => (
                      <span className="chip chip-neutral" key={source}>{source}: {count}</span>
                    ))}
                </div>
              </div>
            </article>
          </section>

          <section className="dashboard-detail-layout">
            <section className="panel">
              <div className="panel-header">
                <div>
                  <h2>Devices</h2>
                  <p>Latest observation per device, including how each host was detected.</p>
                </div>
                <span>{devices.length} records</span>
              </div>

              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>IP</th>
                      <th>Name</th>
                      <th>Source</th>
                      <th>Type</th>
                      <th>Vendor</th>
                      <th>MAC</th>
                      <th>Status</th>
                      <th>Evidence</th>
                      <th>Services</th>
                    </tr>
                  </thead>
                  <tbody>
                    {devices.length === 0 ? (
                      <tr>
                        <td colSpan="9" className="empty-row">No devices discovered yet.</td>
                      </tr>
                    ) : (
                      devices.map((device) => (
                        <tr data-testid={`device-row-${device.id}`} key={device.id} className="clickable-row" onClick={() => handleDeviceSelect(device.id)}>
                          <td>{device.ip}</td>
                          <td>{device.displayName || device.hostname || device.ip}</td>
                          <td>{device.nameSource || '-'}</td>
                          <td>{device.deviceType || '-'}</td>
                          <td>{device.vendor || '-'}</td>
                          <td>{device.mac || '-'}</td>
                          <td>
                            <span className={device.online ? 'chip chip-online' : 'chip chip-offline'}>
                              {device.online ? 'online' : 'offline'}
                            </span>
                          </td>
                          <td>
                            <div className="evidence-group evidence-cell">
                              {device.evidence?.length
                                ? device.evidence.map((source) => <span className="chip chip-neutral" key={source}>{source}</span>)
                                : '-'}
                            </div>
                          </td>
                          <td>
                            {device.services?.length
                              ? device.services.map((service) => service.port).join(', ')
                              : '-'}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </section>

            <DeviceDetailPanel
              detail={selectedDeviceDetail}
              loading={detailLoading}
              error={detailError}
              onClose={() => setSelectedDeviceId(null)}
            />
          </section>
        </>
      ) : (
        <section className="topology-layout">
          <article className="panel topology-panel">
            <div className="panel-header">
              <div>
                <h2>Topology Graph</h2>
                <p>Gateway-centered layout derived from the latest completed scan.</p>
              </div>
              <span>{topology.nodes?.length || 0} nodes</span>
            </div>

            <div className="topology-controls">
              <input
                data-testid="topology-search"
                className="search-input"
                type="search"
                value={filters.search}
                placeholder="Search hostname, IP, vendor"
                onChange={(event) => setFilters((current) => ({ ...current, search: event.target.value }))}
              />
              <label data-testid="filter-online-only" className="toggle-row">
                <input
                  type="checkbox"
                  checked={filters.onlineOnly}
                  onChange={(event) => setFilters((current) => ({ ...current, onlineOnly: event.target.checked }))}
                />
                Online only
              </label>
              <label data-testid="filter-services-only" className="toggle-row">
                <input
                  type="checkbox"
                  checked={filters.onlyServices}
                  onChange={(event) => setFilters((current) => ({ ...current, onlyServices: event.target.checked }))}
                />
                Services only
              </label>
              <select
                data-testid="filter-evidence"
                className="filter-select"
                value={filters.evidence}
                onChange={(event) => setFilters((current) => ({ ...current, evidence: event.target.value }))}
              >
                {evidenceOptions.map((option) => (
                  <option key={option} value={option}>{option === 'all' ? 'All evidence' : option}</option>
                ))}
              </select>
              <select
                data-testid="filter-device-type"
                className="filter-select"
                value={filters.deviceType}
                onChange={(event) => setFilters((current) => ({ ...current, deviceType: event.target.value }))}
              >
                {deviceTypeOptions.map((option) => (
                  <option key={option} value={option}>{option === 'all' ? 'All device types' : option}</option>
                ))}
              </select>
              <select
                data-testid="filter-confidence"
                className="filter-select"
                value={filters.confidence}
                onChange={(event) => setFilters((current) => ({ ...current, confidence: event.target.value }))}
              >
                {confidenceOptions.map((option) => (
                  <option key={option} value={option}>{option === 'all' ? 'All confidence' : formatLabel(option)}</option>
                ))}
              </select>
              <select
                data-testid="filter-name-source"
                className="filter-select"
                value={filters.nameSource}
                onChange={(event) => setFilters((current) => ({ ...current, nameSource: event.target.value }))}
              >
                {nameSourceOptions.map((option) => (
                  <option key={option} value={option}>{option === 'all' ? 'All name sources' : option}</option>
                ))}
              </select>
              <select
                data-testid="filter-layout"
                className="filter-select"
                value={layoutName}
                onChange={(event) => setLayoutName(event.target.value)}
              >
                {layoutOptions.map((option) => (
                  <option key={option.value} value={option.value}>{option.label}</option>
                ))}
              </select>
              <button data-testid="reset-focus" className="secondary-button" onClick={() => setSelectedNode(null)}>Reset Focus</button>
            </div>

            <div className="topology-legend" data-testid="topology-legend">
              <div className="legend-section">
                <strong>Device Types</strong>
                <div className="legend-list">
                  {[
                    ['Gateway', 'legend-gateway'],
                    ['Router/AP', 'legend-router'],
                    ['Computer', 'legend-computer'],
                    ['NAS', 'legend-nas'],
                    ['Printer', 'legend-printer'],
                    ['Media', 'legend-media'],
                    ['Unknown', 'legend-unknown'],
                  ].map(([label, className]) => (
                    <span className="legend-item" key={label}><span className={`legend-swatch ${className}`} />{label}</span>
                  ))}
                </div>
              </div>
              <div className="legend-section">
                <strong>Confidence</strong>
                <div className="legend-list">
                  <span className="legend-item"><span className="legend-ring legend-high" />High</span>
                  <span className="legend-item"><span className="legend-ring legend-medium" />Medium</span>
                  <span className="legend-item"><span className="legend-ring legend-low" />Low</span>
                </div>
              </div>
            </div>

            <TopologyGraph
              topology={topology}
              selectedNodeId={selectedNode?.id}
              onSelectNode={handleNodeSelect}
              filters={filters}
              layoutName={layoutName}
            />
          </article>

          <DeviceDetailPanel
            detail={selectedDeviceDetail}
            loading={detailLoading}
            error={detailError}
            title={selectedNode?.label ? `${selectedNode.label}` : 'Node Details'}
            onClose={() => {
              setSelectedDeviceId(null)
              setSelectedNode(null)
            }}
          />
        </section>
      )}
    </div>
  )
}

function formatTimestamp(value) {
  if (!value) {
    return 'Never'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return 'Unknown'
  }
  return date.toLocaleString()
}

function formatDuration(value) {
  if (!value) {
    return '0 ms'
  }
  if (value < 1000) {
    return `${value} ms`
  }
  return `${(value / 1000).toFixed(1)} s`
}

function formatLabel(value) {
  return value.charAt(0).toUpperCase() + value.slice(1)
}

export default App
