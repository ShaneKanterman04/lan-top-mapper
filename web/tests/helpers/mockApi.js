import {
  summaryFixture,
  devicesFixture,
  topologyFixture,
  deviceDetailFixture,
  scanStartFixture,
  scanStatusRunningFixture,
} from '../fixtures/api-responses.js'

export async function setupMockApi(page, overrides = {}) {
  const routes = {
    summary: overrides.summary !== undefined ? overrides.summary : summaryFixture,
    devices: overrides.devices !== undefined ? overrides.devices : devicesFixture,
    topology: overrides.topology !== undefined ? overrides.topology : topologyFixture,
    deviceDetail: overrides.deviceDetail !== undefined ? overrides.deviceDetail : deviceDetailFixture,
    scanStart: overrides.scanStart !== undefined ? overrides.scanStart : scanStartFixture,
    scanStatus: overrides.scanStatus !== undefined ? overrides.scanStatus : scanStatusRunningFixture,
  }

  await page.route('/api/summary', async (route) => {
    if (routes.summary === null) {
      return route.fulfill({ status: 500, body: JSON.stringify({ error: 'Summary failed' }) })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(routes.summary),
    })
  })

  await page.route('/api/devices', async (route) => {
    if (routes.devices === null) {
      return route.fulfill({ status: 500, body: JSON.stringify({ error: 'Devices failed' }) })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(routes.devices),
    })
  })

  await page.route('/api/topology', async (route) => {
    if (routes.topology === null) {
      return route.fulfill({ status: 500, body: JSON.stringify({ error: 'Topology failed' }) })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(routes.topology),
    })
  })

  await page.route(/\/api\/devices\/\d+/, async (route) => {
    if (routes.deviceDetail === null) {
      return route.fulfill({ status: 500, body: JSON.stringify({ error: 'Device detail failed' }) })
    }
    const url = route.request().url()
    const match = url.match(/\/api\/devices\/(\d+)/)
    const id = match ? parseInt(match[1], 10) : 0
    const detail = typeof routes.deviceDetail === 'function'
      ? routes.deviceDetail(id)
      : routes.deviceDetail
    if (!detail) {
      return route.fulfill({ status: 404, body: JSON.stringify({ error: 'device not found' }) })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(detail),
    })
  })

  await page.route('/api/scans', async (route) => {
    if (route.request().method() !== 'POST') {
      return route.fulfill({ status: 405, body: JSON.stringify({ error: 'method not allowed' }) })
    }
    if (routes.scanStart === null) {
      return route.fulfill({ status: 500, body: JSON.stringify({ error: 'Scan start failed' }) })
    }
    return route.fulfill({
      status: 202,
      contentType: 'application/json',
      body: JSON.stringify(routes.scanStart),
    })
  })

  await page.route(/\/api\/scans\/\d+/, async (route) => {
    if (route.request().method() !== 'GET') {
      return route.fulfill({ status: 405, body: JSON.stringify({ error: 'method not allowed' }) })
    }
    if (routes.scanStatus === null) {
      return route.fulfill({ status: 500, body: JSON.stringify({ error: 'Scan status failed' }) })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(routes.scanStatus),
    })
  })
}
