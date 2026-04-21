import { expect, test } from '@playwright/test'
import { setupMockApi } from './helpers/mockApi.js'
import { devicesFixture, deviceDetailFixture } from './fixtures/api-responses.js'

test.describe('device detail panel', () => {
  test.beforeEach(async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })
  })

  test('shows empty prompt before selection', async ({ page }) => {
    await expect(page.getByTestId('device-detail-empty')).toBeVisible()
    await expect(page.getByTestId('device-detail-empty')).toContainText('Select a device to inspect its history.')
  })

  test('shows loading state while fetching detail', async ({ page }) => {
    await page.route('/api/devices/**', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 500))
      return route.continue()
    })

    const firstDevice = devicesFixture[0]
    await page.getByTestId(`device-row-${firstDevice.id}`).click()

    await expect(page.getByTestId('device-detail-loading')).toBeVisible()
  })

  test('shows error state when detail fetch fails', async ({ page }) => {
    await setupMockApi(page, { deviceDetail: null })
    await page.reload()

    const firstDevice = devicesFixture[0]
    await page.getByTestId(`device-row-${firstDevice.id}`).click()

    await expect(page.getByTestId('device-detail-error')).toBeVisible()
    await expect(page.getByTestId('device-detail-error')).toContainText('Device detail failed')
  })

  test('renders all detail sections for selected device', async ({ page }) => {
    const firstDevice = devicesFixture[0]
    await page.getByTestId(`device-row-${firstDevice.id}`).click()

    await expect(page.getByTestId('device-detail-panel')).toBeVisible()
    await expect(page.getByText('Overview')).toBeVisible()
    await expect(page.getByText('Classification')).toBeVisible()
    await expect(page.getByText('Identity Claims')).toBeVisible()
    await expect(page.getByText('Current Evidence')).toBeVisible()
    await expect(page.getByText('Latest Services')).toBeVisible()
    await expect(page.getByText('Recent Observations')).toBeVisible()
  })

  test('shows correct device overview data', async ({ page }) => {
    const device = devicesFixture[0]
    await page.getByTestId(`device-row-${device.id}`).click()

    await expect(page.getByTestId('device-detail-panel')).toContainText(device.displayName)
    await expect(page.getByTestId('device-detail-panel')).toContainText(device.ip)
    await expect(page.getByTestId('device-detail-panel')).toContainText(device.mac)
    await expect(page.getByTestId('device-detail-panel')).toContainText(device.vendor)
  })

  test('close button clears the panel', async ({ page }) => {
    const firstDevice = devicesFixture[0]
    await page.getByTestId(`device-row-${firstDevice.id}`).click()

    await expect(page.getByTestId('device-detail-panel')).toBeVisible()
    await page.getByRole('button', { name: 'Close' }).click()

    await expect(page.getByTestId('device-detail-empty')).toBeVisible()
  })

  test('shows device type and classification', async ({ page }) => {
    const device = devicesFixture[0]
    await page.getByTestId(`device-row-${device.id}`).click()

    await expect(page.getByTestId('device-detail-panel')).toContainText(device.deviceType)
    await expect(page.getByTestId('device-detail-panel')).toContainText(device.classificationConfidence)
  })

  test('renders observation history', async ({ page }) => {
    const device = devicesFixture[0]
    await page.getByTestId(`device-row-${device.id}`).click()

    await expect(page.getByText('Recent Observations')).toBeVisible()
    const detail = deviceDetailFixture(device.id)
    if (detail.observationHistory.length > 0) {
      await expect(page.getByText(`Scan #${detail.observationHistory[0].scanId}`)).toBeVisible()
    }
  })
})
