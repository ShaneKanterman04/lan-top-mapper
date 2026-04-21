import { expect, test } from '@playwright/test'
import { setupMockApi } from './helpers/mockApi.js'
import { emptyDevicesFixture, emptyTopologyFixture, emptySummaryFixture } from './fixtures/api-responses.js'

test.describe('app shell', () => {
  test('renders dashboard by default with mocked data', async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })

    await expect(page.getByTestId('tab-dashboard')).toBeVisible()
    await expect(page.getByTestId('tab-topology')).toBeVisible()
    await expect(page.getByTestId('metric-card-subnet')).toBeVisible()
    await expect(page.getByTestId('metric-card-devices')).toBeVisible()
    await expect(page.getByTestId('metric-card-online')).toBeVisible()
    await expect(page.getByTestId('metric-card-mode')).toBeVisible()
  })

  test('switches to topology tab', async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })

    await page.getByTestId('tab-topology').click()
    await expect(page.getByTestId('topology-canvas')).toBeVisible()
    await expect(page.getByTestId('topology-legend')).toBeVisible()
  })

  test('shows global error when summary fails to load', async ({ page }) => {
    await setupMockApi(page, { summary: null })
    await page.goto('/', { waitUntil: 'networkidle' })

    await expect(page.getByTestId('error-box')).toBeVisible()
    await expect(page.getByTestId('error-box')).toContainText('Summary failed')
  })

  test('shows empty state when no devices discovered', async ({ page }) => {
    await setupMockApi(page, {
      summary: emptySummaryFixture,
      devices: emptyDevicesFixture,
      topology: emptyTopologyFixture,
    })
    await page.goto('/', { waitUntil: 'networkidle' })

    await expect(page.getByText('No devices discovered yet.')).toBeVisible()
  })

  test('shows correct scan status pill', async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })

    await expect(page.getByTestId('status-pill')).toHaveText('idle')
  })

  test('start scan button is enabled when idle', async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })

    await expect(page.getByTestId('start-scan-button')).toBeEnabled()
    await expect(page.getByTestId('start-scan-button')).toHaveText('Start Scan')
  })
})
