import { expect, test } from '@playwright/test'
import { setupMockApi } from './helpers/mockApi.js'
import { topologyFixture, emptyTopologyFixture } from './fixtures/api-responses.js'

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

test.describe('topology view', () => {
  test.beforeEach(async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })
    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
    await page.waitForFunction(() => Boolean(window.__LTM_TOPOLOGY_CY__))
  })

  test('renders topology graph and legend', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Topology Graph' })).toBeVisible()
    await expect(page.getByTestId('topology-legend')).toBeVisible()
  })

  test('switches layout modes', async ({ page }) => {
    await page.getByTestId('filter-layout').selectOption('concentric')
    await page.waitForTimeout(400)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()

    await page.getByTestId('filter-layout').selectOption('breadthfirst')
    await page.waitForTimeout(400)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()

    await page.getByTestId('filter-layout').selectOption('radial')
    await page.waitForTimeout(400)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('filters by device type', async ({ page }) => {
    await page.getByTestId('filter-device-type').selectOption('Computer')
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('filters by confidence', async ({ page }) => {
    await page.getByTestId('filter-confidence').selectOption('high')
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('filters by name source', async ({ page }) => {
    await page.getByTestId('filter-name-source').selectOption('dns')
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('filters by evidence', async ({ page }) => {
    await page.getByTestId('filter-evidence').selectOption('arp')
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('search filters topology', async ({ page }) => {
    await page.getByTestId('topology-search').fill('Printer')
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('online only toggle filters topology', async ({ page }) => {
    await page.getByTestId('filter-online-only').locator('input').check()
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('services only toggle filters topology', async ({ page }) => {
    await page.getByTestId('filter-services-only').locator('input').check()
    await page.waitForTimeout(300)
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })

  test('selects gateway and updates detail panel', async ({ page }) => {
    await clickNode(page, 'gateway')
    await expect(page.getByText('Name Source', { exact: true })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Identity Claims' })).toBeVisible()
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
  })

  test('reset focus clears selected node', async ({ page }) => {
    await clickNode(page, 'gateway')
    await expect(page.getByText('Name Source', { exact: true })).toBeVisible()

    await page.getByTestId('reset-focus').click()
    await page.waitForTimeout(300)
    await expect(page.getByTestId('device-detail-empty')).toBeVisible()
  })

  test('renders empty topology gracefully', async ({ page }) => {
    await setupMockApi(page, { topology: emptyTopologyFixture })
    await page.reload()
    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
  })
})
