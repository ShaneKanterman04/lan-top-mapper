import { expect, test } from '@playwright/test'
import { setupMockApi } from './helpers/mockApi.js'
import {
  summaryFixture,
  scanStartFixture,
  scanStatusRunningFixture,
  scanStatusCompletedFixture,
  scanStatusFailedFixture,
} from './fixtures/api-responses.js'

test.describe('scan flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle' })
  })

  test('start scan button submits scan and shows running state', async ({ page }) => {
    await setupMockApi(page, {
      summary: { ...summaryFixture, latestScanState: 'idle' },
      scanStart: scanStartFixture,
      scanStatus: scanStatusRunningFixture,
    })
    await page.reload()

    await expect(page.getByTestId('start-scan-button')).toBeEnabled()
    await expect(page.getByTestId('status-pill')).toHaveText('idle')

    await page.getByTestId('start-scan-button').click()

    await expect(page.getByTestId('start-scan-button')).toBeDisabled()
    await expect(page.getByTestId('start-scan-button')).toHaveText('Scanning...')
    await expect(page.getByTestId('status-pill')).toHaveText('running')
  })

  test('scan completes and updates status', async ({ page }) => {
    let pollCount = 0
    await setupMockApi(page, {
      summary: { ...summaryFixture, latestScanState: 'idle' },
      scanStart: scanStartFixture,
      scanStatus: scanStatusCompletedFixture,
    })

    await page.route(/\/api\/scans\/\d+/, async (route) => {
      pollCount++
      const status = pollCount <= 2 ? scanStatusRunningFixture : scanStatusCompletedFixture
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(status),
      })
    })

    await page.reload()
    await page.getByTestId('start-scan-button').click()

    await page.waitForTimeout(4000)
    await expect(page.getByTestId('status-pill')).toHaveText('idle')
  })

  test('scan start failure shows error', async ({ page }) => {
    await setupMockApi(page, {
      summary: { ...summaryFixture, latestScanState: 'idle' },
      scanStart: null,
    })
    await page.reload()

    await page.getByTestId('start-scan-button').click()

    await expect(page.getByTestId('error-box')).toBeVisible()
    await expect(page.getByTestId('error-box')).toContainText('Scan start failed')
  })

  test('scan fails and updates status', async ({ page }) => {
    await setupMockApi(page, {
      summary: { ...summaryFixture, latestScanState: 'idle' },
      scanStart: scanStartFixture,
      scanStatus: scanStatusFailedFixture,
    })
    await page.reload()

    await page.getByTestId('start-scan-button').click()
    await page.waitForTimeout(2000)

    await expect(page.getByTestId('status-pill')).toHaveText('failed')
  })

  test('button disabled when scan already running', async ({ page }) => {
    await setupMockApi(page, {
      summary: { ...summaryFixture, latestScanState: 'running' },
    })
    await page.reload()

    await expect(page.getByTestId('start-scan-button')).toBeDisabled()
    await expect(page.getByTestId('start-scan-button')).toHaveText('Scanning...')
    await expect(page.getByTestId('status-pill')).toHaveText('running')
  })
})
