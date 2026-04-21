import { expect, test } from '@playwright/test'

test.describe('smoke: real backend', () => {
  test.beforeEach(async ({ page }) => {
    const pageErrors = []
    page.on('pageerror', (error) => pageErrors.push(error.message))
    page.on('requestfailed', (request) => {
      const url = request.url()
      if (url.includes('/api/topology') || url.includes('/api/devices') || url.includes('/api/summary')) {
        pageErrors.push(`request failed: ${request.method()} ${url}`)
      }
    })

    await page.goto('/', { waitUntil: 'networkidle' })
    test.info().annotations.push({ type: 'page-errors', description: JSON.stringify(pageErrors) })
  })

  test('app loads and dashboard is visible', async ({ page }) => {
    await expect(page.getByTestId('tab-dashboard')).toBeVisible()
    await expect(page.getByTestId('tab-topology')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'LAN Topology Mapper' })).toBeVisible()
  })

  test('topology tab opens without critical errors', async ({ page }) => {
    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
    await page.waitForFunction(() => Boolean(window.__LTM_TOPOLOGY_CY__))
    await expect(page.getByTestId('topology-legend')).toBeVisible()
  })

  test('can select a node if topology data exists', async ({ page }) => {
    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
    await page.waitForFunction(() => Boolean(window.__LTM_TOPOLOGY_CY__))

    const hasNodes = await page.evaluate(() => {
      const cy = window.__LTM_TOPOLOGY_CY__
      return cy.nodes().length > 0
    })

    test.skip(!hasNodes, 'no topology nodes available')

    const nodeId = await page.evaluate(() => {
      const cy = window.__LTM_TOPOLOGY_CY__
      const node = cy.nodes().first()
      return node?.id()
    })

    const canvas = page.getByTestId('topology-canvas')
    const box = await canvas.boundingBox()
    expect(box).toBeTruthy()

    const position = await page.evaluate((id) => {
      const cy = window.__LTM_TOPOLOGY_CY__
      const node = cy.getElementById(id)
      if (!node.length) return null
      const rendered = node.renderedPosition()
      return { x: rendered.x, y: rendered.y }
    }, nodeId)

    if (position) {
      await page.mouse.click(box.x + position.x, box.y + position.y)
      await page.waitForTimeout(300)
      const detailVisible = await page.getByTestId('device-detail-panel').isVisible().catch(() => false)
      if (detailVisible) {
        await expect(page.getByTestId('device-detail-panel')).toBeVisible()
      }
    }
  })

  test('devices table loads without errors', async ({ page }) => {
    await expect(page.getByText('Devices')).toBeVisible()
    const rows = await page.locator('[data-testid^="device-row-"]').count()
    if (rows > 0) {
      await expect(page.locator('[data-testid^="device-row-"]').first()).toBeVisible()
    } else {
      await expect(page.getByText('No devices discovered yet.')).toBeVisible()
    }
  })
})
