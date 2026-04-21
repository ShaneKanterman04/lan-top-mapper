import { expect, test } from '@playwright/test'
import { setupMockApi } from './helpers/mockApi.js'
import { summaryFixture, devicesFixture, emptyDevicesFixture, emptySummaryFixture } from './fixtures/api-responses.js'

test.describe('dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await setupMockApi(page)
    await page.goto('/', { waitUntil: 'networkidle' })
  })

  test('renders metric cards with correct values', async ({ page }) => {
    await expect(page.getByTestId('metric-card-subnet')).toContainText('192.168.1.0/24')
    await expect(page.getByTestId('metric-card-devices')).toContainText('5')
    await expect(page.getByTestId('metric-card-online')).toContainText('4')
    await expect(page.getByTestId('metric-card-mode')).toContainText('Enhanced')
  })

  test('renders scan snapshot details', async ({ page }) => {
    await expect(page.getByText('Gateway', { exact: true })).toBeVisible()
    await expect(page.getByText('192.168.1.1')).toBeVisible()
    await expect(page.getByText('Available Sources')).toBeVisible()
    await expect(page.getByText('arp, icmp, tcp, mdns')).toBeVisible()
  })

  test('renders discovery diagnostics', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Discovery Diagnostics' })).toBeVisible()
    await expect(page.getByText('Devices found: 5')).toBeVisible()
    await expect(page.getByText('Duration: 12.5 s')).toBeVisible()
  })

  test('renders devices table with correct rows', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Devices' })).toBeVisible()
    await expect(page.getByText('5 records')).toBeVisible()

    for (const device of devicesFixture) {
      const row = page.getByTestId(`device-row-${device.id}`)
      await expect(row).toBeVisible()
      await expect(row).toContainText(device.ip)
    }
  })

  test('shows empty state when no devices', async ({ page }) => {
    await setupMockApi(page, {
      summary: emptySummaryFixture,
      devices: emptyDevicesFixture,
    })
    await page.reload()

    await expect(page.getByText('No devices discovered yet.')).toBeVisible()
  })

  test('clicking device row opens detail panel', async ({ page }) => {
    const firstDevice = devicesFixture[0]
    await page.getByTestId(`device-row-${firstDevice.id}`).click()

    await expect(page.getByTestId('device-detail-panel')).toBeVisible()
    await expect(page.getByTestId('device-detail-panel')).toContainText(firstDevice.displayName || firstDevice.ip)
  })

  test('device row shows correct status chips', async ({ page }) => {
    const onlineDevice = devicesFixture[0]
    const offlineDevice = devicesFixture[2]

    await expect(page.getByTestId(`device-row-${onlineDevice.id}`)).toContainText('online')
    await expect(page.getByTestId(`device-row-${offlineDevice.id}`)).toContainText('offline')
  })
})
