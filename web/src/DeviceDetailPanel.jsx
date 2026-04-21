import React from 'react'

function DeviceDetailPanel({ detail, loading, error, title = 'Device Details', onClose }) {
  return (
    <article data-testid="device-detail-panel" className="panel detail-panel device-detail-panel">
      <div className="panel-header">
        <div>
          <h2>{title}</h2>
          <p>Current identity, latest observation, and recent scan history.</p>
        </div>
        {onClose ? <button className="secondary-button" onClick={onClose}>Close</button> : null}
      </div>

      {loading ? <p data-testid="device-detail-loading">Loading device details...</p> : null}
      {error ? <section data-testid="device-detail-error" className="error-box">{error}</section> : null}
      {!loading && !error && !detail ? <p data-testid="device-detail-empty">Select a device to inspect its history.</p> : null}

      {!loading && !error && detail ? (
        <div className="detail-stack">
          <section className="detail-section">
            <h3>Overview</h3>
            <div className="detail-grid">
              <DetailItem label="Display Name" value={detail.device.displayName || '-'} />
              <DetailItem label="Name Source" value={detail.device.nameSource || '-'} />
              <DetailItem label="Device Type" value={detail.device.deviceType || '-'} />
              <DetailItem label="Type Source" value={detail.device.typeSource || '-'} />
              <DetailItem label="Confidence" value={detail.device.classificationConfidence || '-'} />
              <DetailItem label="IP" value={detail.device.ip} />
              <DetailItem label="Hostname" value={detail.device.hostname || '-'} />
              <DetailItem label="MAC" value={detail.device.mac || '-'} />
              <DetailItem label="Vendor" value={detail.device.vendor || '-'} />
              <DetailItem label="Vendor Source" value={detail.device.vendorSource || '-'} />
              <DetailItem label="First Seen" value={formatTimestamp(detail.device.firstSeen)} />
              <DetailItem label="Last Seen" value={formatTimestamp(detail.device.lastSeen)} />
              <div>
                <span className="detail-label">Status</span>
                <span className={detail.device.online ? 'chip chip-online' : 'chip chip-offline'}>
                  {detail.device.online ? 'online' : 'offline'}
                </span>
              </div>
              <DetailItem label="Times Seen" value={detail.scanSummary.timesSeen} />
            </div>
          </section>

          <section className="detail-section">
            <h3>Classification</h3>
            <div className="evidence-group">
              {detail.device.classificationReasons?.length
                ? detail.device.classificationReasons.map((reason) => <span className="chip chip-neutral" key={reason}>{reason}</span>)
                : <span>-</span>}
            </div>
          </section>

          <section className="detail-section">
            <h3>Identity Claims</h3>
            <div className="history-stack">
              {detail.device.identityClaims?.length
                ? detail.device.identityClaims.map((claim, index) => (
                  <article className="history-card" key={`${claim.source}-${claim.claimType}-${claim.value}-${index}`}>
                    <div className="history-meta">
                      <span>{claim.source}</span>
                      <span>{claim.claimType}</span>
                      <span>{claim.confidenceLabel || 'unscored'}</span>
                    </div>
                    <strong>{claim.value}</strong>
                  </article>
                ))
                : <p>No identity claims recorded yet.</p>}
            </div>
          </section>

          <section className="detail-section">
            <h3>Current Evidence</h3>
            <div className="evidence-group">
              {detail.device.evidence?.length
                ? detail.device.evidence.map((source) => <span className="chip chip-neutral" key={source}>{source}</span>)
                : <span>-</span>}
            </div>
            <div className="summary-metric-row">
              {Object.entries(detail.evidenceSummary || {}).map(([source, count]) => (
                <span className="chip chip-neutral" key={source}>{source}: {count}</span>
              ))}
            </div>
          </section>

          <section className="detail-section">
            <h3>Latest Services</h3>
            <div className="evidence-group">
              {detail.device.services?.length
                ? detail.device.services.map((service) => (
                  <span className="chip chip-neutral" key={`${service.protocol}-${service.port}`}>
                    {service.serviceName} ({service.port})
                  </span>
                ))
                : <span>-</span>}
            </div>
          </section>

          <section className="detail-section">
            <h3>Recent Observations</h3>
            <div className="history-stack">
              {detail.observationHistory?.length ? detail.observationHistory.map((item) => (
                <article className="history-card" key={item.observationId}>
                  <div className="history-header">
                    <strong>{formatTimestamp(item.observedAt)}</strong>
                    <span className={item.online ? 'chip chip-online' : 'chip chip-offline'}>
                      {item.online ? 'online' : 'offline'}
                    </span>
                  </div>
                  <div className="history-meta">
                    <span>Scan #{item.scanId}</span>
                    <span>{item.scanStatus || 'unknown'}</span>
                    <span>{item.displayName || item.hostname || item.ip}</span>
                    <span>{item.deviceType || 'Unknown device'}</span>
                  </div>
                  <div className="evidence-group">
                    {item.classificationReasons?.length
                      ? item.classificationReasons.map((reason) => <span className="chip chip-neutral" key={`${item.observationId}-reason-${reason}`}>{reason}</span>)
                      : <span className="muted-text">No classification clues recorded</span>}
                  </div>
                  <div className="evidence-group">
                    {item.evidence?.length
                      ? item.evidence.map((source) => <span className="chip chip-neutral" key={`${item.observationId}-${source}`}>{source}</span>)
                      : <span>-</span>}
                  </div>
                  <div className="evidence-group">
                    {item.services?.length
                      ? item.services.map((service) => (
                        <span className="chip chip-neutral" key={`${item.observationId}-${service.protocol}-${service.port}`}>
                          {service.serviceName} ({service.port})
                        </span>
                      ))
                      : <span className="muted-text">No services recorded</span>}
                  </div>
                </article>
              )) : <p>No observation history yet.</p>}
            </div>
          </section>
        </div>
      ) : null}
    </article>
  )
}

function DetailItem({ label, value }) {
  return (
    <div>
      <span className="detail-label">{label}</span>
      <strong>{value}</strong>
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

export default DeviceDetailPanel
