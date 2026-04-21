import { expect, test } from '@playwright/test'

test.describe('topology ux', () => {
  test.beforeEach(async ({ page }) => {
    const pageErrors = []
    page.on('pageerror', (error) => pageErrors.push(error.message))
    page.on('requestfailed', (request) => {
      const url = request.url()
      if (url.includes('/api/topology') || url.includes('/api/devices') || url.includes('/api/summary')) {
        pageErrors.push(`request failed: ${request.method()} ${url}`)
      }
    })

    await page.goto('/')
    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
    await page.waitForFunction(() => Boolean(window.__LTM_TOPOLOGY_CY__))

    test.info().annotations.push({ type: 'page-errors', description: JSON.stringify(pageErrors) })
  })

  test('renders topology graph and legend', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Topology Graph' })).toBeVisible()
    await expect(page.getByTestId('topology-legend')).toBeVisible()
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-desktop-default.png') })
  })

  test('switches layout modes', async ({ page }) => {
    await page.getByTestId('filter-layout').selectOption('concentric')
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-desktop-concentric.png') })

    await page.getByTestId('filter-layout').selectOption('breadthfirst')
    await page.waitForTimeout(400)
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-desktop-breadthfirst.png') })

    await page.getByTestId('filter-layout').selectOption('radial')
    await page.waitForTimeout(400)
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-desktop-radial.png') })
  })

  test('filters by low confidence and unknown device types', async ({ page }) => {
    await page.getByTestId('filter-confidence').selectOption('low')
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-filter-low-confidence.png') })

    await page.getByTestId('filter-device-type').selectOption('Unknown device')
    await page.waitForTimeout(300)
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-filter-unknown-devices.png') })
  })

  test('filters by name source', async ({ page }) => {
    await page.getByTestId('filter-name-source').selectOption('upnp-desc')
    await page.waitForTimeout(300)
    await page.getByTestId('topology-canvas').screenshot({ path: test.info().outputPath('topology-filter-upnp-source.png') })
  })

  test('selects gateway and updates detail panel', async ({ page }) => {
    await clickNode(page, 'gateway')
    await expect(page.getByText('Name Source', { exact: true })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Identity Claims' })).toBeVisible()
    await page.screenshot({ path: test.info().outputPath('topology-selected-gateway.png'), fullPage: true })
  })

  test('selects a non-gateway node and updates detail panel', async ({ page }) => {
    const nodeId = await page.evaluate(() => {
      const cy = window.__LTM_TOPOLOGY_CY__
      const node = cy.nodes().filter((item) => item.data('kind') !== 'gateway').first()
      return node?.id()
    })
    expect(nodeId).toBeTruthy()
    await clickNode(page, nodeId)
    await expect(page.getByText('Recent Observations')).toBeVisible()
    await page.screenshot({ path: test.info().outputPath('topology-selected-node.png'), fullPage: true })
  })

  test('renders on mobile', async ({ browser }) => {
    const page = await browser.newPage({ viewport: { width: 390, height: 844 } })
    await page.goto('http://127.0.0.1:8080')
    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
    await page.screenshot({ path: test.info().outputPath('topology-mobile-default.png'), fullPage: true })
    await page.close()
  })
})

async function clickNode(page, nodeId) {
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

  expect(position).toBeTruthy()
  await page.mouse.click(box.x + position.x, box.y + position.y)
  await page.waitForTimeout(300)
}
