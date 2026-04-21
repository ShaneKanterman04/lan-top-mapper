# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: dashboard.spec.js >> dashboard >> renders scan snapshot details
- Location: tests/dashboard.spec.js:18:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByText('Gateway')
Expected: visible
Error: strict mode violation: getByText('Gateway') resolved to 2 elements:
    1) <dt>Gateway</dt> aka getByText('Gateway', { exact: true })
    2) <td>Gateway Router</td> aka getByRole('cell', { name: 'Gateway Router' })

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByText('Gateway')

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - generic [ref=e5]:
      - paragraph [ref=e6]: Linux Discovery Diagnostics
      - heading "LAN Topology Mapper" [level=1] [ref=e7]
      - paragraph [ref=e8]: Run repeatable LAN scans, inspect discovery evidence, and explore the current topology graph.
    - generic [ref=e9]:
      - button "Start Scan" [ref=e10] [cursor=pointer]
      - generic [ref=e11]: idle
  - navigation [ref=e12]:
    - button "Dashboard" [ref=e13] [cursor=pointer]
    - button "Topology" [ref=e14] [cursor=pointer]
  - generic [ref=e15]: "Enhanced mode: raw active ARP is the primary LAN discovery method, with ICMP, TCP, and neighbor-cache data used for enrichment."
  - generic [ref=e16]:
    - article [ref=e17]:
      - generic [ref=e18]: Subnet
      - strong [ref=e19]: 192.168.1.0/24
    - article [ref=e20]:
      - generic [ref=e21]: Devices
      - strong [ref=e22]: "5"
    - article [ref=e23]:
      - generic [ref=e24]: Online
      - strong [ref=e25]: "4"
    - article [ref=e26]:
      - generic [ref=e27]: Mode
      - strong [ref=e28]: Enhanced
  - generic [ref=e29]:
    - article [ref=e30]:
      - heading "Scan Snapshot" [level=2] [ref=e31]
      - generic [ref=e32]:
        - generic [ref=e33]:
          - term [ref=e34]: Gateway
          - definition [ref=e35]: 192.168.1.1
        - generic [ref=e36]:
          - term [ref=e37]: Last Scan
          - definition [ref=e38]: 4/21/2026, 9:43:19 PM
        - generic [ref=e39]:
          - term [ref=e40]: Available Sources
          - definition [ref=e41]: arp, icmp, tcp, mdns
        - generic [ref=e42]:
          - term [ref=e43]: Targets
          - definition [ref=e44]: "254"
    - article [ref=e45]:
      - heading "Discovery Diagnostics" [level=2] [ref=e46]
      - generic [ref=e47]:
        - paragraph [ref=e48]:
          - text: "Devices found:"
          - strong [ref=e49]: "5"
        - paragraph [ref=e50]:
          - text: "Duration:"
          - strong [ref=e51]: 12.5 s
        - generic [ref=e52]:
          - generic [ref=e53]: "arp: 3"
          - generic [ref=e54]: "icmp: 2"
          - generic [ref=e55]: "tcp: 1"
  - generic [ref=e56]:
    - generic [ref=e57]:
      - generic [ref=e58]:
        - generic [ref=e59]:
          - heading "Devices" [level=2] [ref=e60]
          - paragraph [ref=e61]: Latest observation per device, including how each host was detected.
        - generic [ref=e62]: 5 records
      - table [ref=e64]:
        - rowgroup [ref=e65]:
          - row "IP Name Source Type Vendor MAC Status Evidence Services" [ref=e66]:
            - columnheader "IP" [ref=e67]
            - columnheader "Name" [ref=e68]
            - columnheader "Source" [ref=e69]
            - columnheader "Type" [ref=e70]
            - columnheader "Vendor" [ref=e71]
            - columnheader "MAC" [ref=e72]
            - columnheader "Status" [ref=e73]
            - columnheader "Evidence" [ref=e74]
            - columnheader "Services" [ref=e75]
        - rowgroup [ref=e76]:
          - row "192.168.1.1 Gateway Router dns Router or access point Netgear AA:BB:CC:11:22:33 online arp icmp 80" [ref=e77] [cursor=pointer]:
            - cell "192.168.1.1" [ref=e78]
            - cell "Gateway Router" [ref=e79]
            - cell "dns" [ref=e80]
            - cell "Router or access point" [ref=e81]
            - cell "Netgear" [ref=e82]
            - cell "AA:BB:CC:11:22:33" [ref=e83]
            - cell "online" [ref=e84]:
              - generic [ref=e85]: online
            - cell "arp icmp" [ref=e86]:
              - generic [ref=e87]:
                - generic [ref=e88]: arp
                - generic [ref=e89]: icmp
            - cell "80" [ref=e90]
          - row "192.168.1.42 192.168.1.42 ip Computer Dell DD:EE:FF:44:55:66 online arp tcp 22, 445" [ref=e91] [cursor=pointer]:
            - cell "192.168.1.42" [ref=e92]
            - cell "192.168.1.42" [ref=e93]
            - cell "ip" [ref=e94]
            - cell "Computer" [ref=e95]
            - cell "Dell" [ref=e96]
            - cell "DD:EE:FF:44:55:66" [ref=e97]
            - cell "online" [ref=e98]:
              - generic [ref=e99]: online
            - cell "arp tcp" [ref=e100]:
              - generic [ref=e101]:
                - generic [ref=e102]: arp
                - generic [ref=e103]: tcp
            - cell "22, 445" [ref=e104]
          - row "192.168.1.100 Living Room TV mdns Media device Samsung 11:22:33:AA:BB:CC offline mdns 8009" [ref=e105] [cursor=pointer]:
            - cell "192.168.1.100" [ref=e106]
            - cell "Living Room TV" [ref=e107]
            - cell "mdns" [ref=e108]
            - cell "Media device" [ref=e109]
            - cell "Samsung" [ref=e110]
            - cell "11:22:33:AA:BB:CC" [ref=e111]
            - cell "offline" [ref=e112]:
              - generic [ref=e113]: offline
            - cell "mdns" [ref=e114]:
              - generic [ref=e116]: mdns
            - cell "8009" [ref=e117]
          - row "192.168.1.55 192.168.1.55 ip - - - online icmp -" [ref=e118] [cursor=pointer]:
            - cell "192.168.1.55" [ref=e119]
            - cell "192.168.1.55" [ref=e120]
            - cell "ip" [ref=e121]
            - cell "-" [ref=e122]
            - cell "-" [ref=e123]
            - cell "-" [ref=e124]
            - cell "online" [ref=e125]:
              - generic [ref=e126]: online
            - cell "icmp" [ref=e127]:
              - generic [ref=e129]: icmp
            - cell "-" [ref=e130]
          - row "192.168.1.10 Printer upnp-desc Printer HP 44:55:66:DD:EE:FF online upnp tcp 9100" [ref=e131] [cursor=pointer]:
            - cell "192.168.1.10" [ref=e132]
            - cell "Printer" [ref=e133]
            - cell "upnp-desc" [ref=e134]
            - cell "Printer" [ref=e135]
            - cell "HP" [ref=e136]
            - cell "44:55:66:DD:EE:FF" [ref=e137]
            - cell "online" [ref=e138]:
              - generic [ref=e139]: online
            - cell "upnp tcp" [ref=e140]:
              - generic [ref=e141]:
                - generic [ref=e142]: upnp
                - generic [ref=e143]: tcp
            - cell "9100" [ref=e144]
    - article [ref=e145]:
      - generic [ref=e146]:
        - generic [ref=e147]:
          - heading "Device Details" [level=2] [ref=e148]
          - paragraph [ref=e149]: Current identity, latest observation, and recent scan history.
        - button "Close" [ref=e150] [cursor=pointer]
      - paragraph [ref=e151]: Select a device to inspect its history.
```

# Test source

```ts
  1  | import { expect, test } from '@playwright/test'
  2  | import { setupMockApi } from './helpers/mockApi.js'
  3  | import { summaryFixture, devicesFixture, emptyDevicesFixture, emptySummaryFixture } from './fixtures/api-responses.js'
  4  | 
  5  | test.describe('dashboard', () => {
  6  |   test.beforeEach(async ({ page }) => {
  7  |     await setupMockApi(page)
  8  |     await page.goto('/', { waitUntil: 'networkidle' })
  9  |   })
  10 | 
  11 |   test('renders metric cards with correct values', async ({ page }) => {
  12 |     await expect(page.getByTestId('metric-card-subnet')).toContainText('192.168.1.0/24')
  13 |     await expect(page.getByTestId('metric-card-devices')).toContainText('5')
  14 |     await expect(page.getByTestId('metric-card-online')).toContainText('4')
  15 |     await expect(page.getByTestId('metric-card-mode')).toContainText('Enhanced')
  16 |   })
  17 | 
  18 |   test('renders scan snapshot details', async ({ page }) => {
> 19 |     await expect(page.getByText('Gateway')).toBeVisible()
     |                                             ^ Error: expect(locator).toBeVisible() failed
  20 |     await expect(page.getByText('192.168.1.1')).toBeVisible()
  21 |     await expect(page.getByText('Available Sources')).toBeVisible()
  22 |     await expect(page.getByText('arp, icmp, tcp, mdns')).toBeVisible()
  23 |   })
  24 | 
  25 |   test('renders discovery diagnostics', async ({ page }) => {
  26 |     await expect(page.getByText('Discovery Diagnostics')).toBeVisible()
  27 |     await expect(page.getByText('Devices found: 5')).toBeVisible()
  28 |     await expect(page.getByText('Duration: 12.5 s')).toBeVisible()
  29 |   })
  30 | 
  31 |   test('renders devices table with correct rows', async ({ page }) => {
  32 |     await expect(page.getByText('Devices')).toBeVisible()
  33 |     await expect(page.getByText('5 records')).toBeVisible()
  34 | 
  35 |     for (const device of devicesFixture) {
  36 |       const row = page.getByTestId(`device-row-${device.id}`)
  37 |       await expect(row).toBeVisible()
  38 |       await expect(row).toContainText(device.ip)
  39 |     }
  40 |   })
  41 | 
  42 |   test('shows empty state when no devices', async ({ page }) => {
  43 |     await setupMockApi(page, {
  44 |       summary: emptySummaryFixture,
  45 |       devices: emptyDevicesFixture,
  46 |     })
  47 |     await page.reload()
  48 | 
  49 |     await expect(page.getByText('No devices discovered yet.')).toBeVisible()
  50 |   })
  51 | 
  52 |   test('clicking device row opens detail panel', async ({ page }) => {
  53 |     const firstDevice = devicesFixture[0]
  54 |     await page.getByTestId(`device-row-${firstDevice.id}`).click()
  55 | 
  56 |     await expect(page.getByTestId('device-detail-panel')).toBeVisible()
  57 |     await expect(page.getByTestId('device-detail-panel')).toContainText(firstDevice.displayName || firstDevice.ip)
  58 |   })
  59 | 
  60 |   test('device row shows correct status chips', async ({ page }) => {
  61 |     const onlineDevice = devicesFixture[0]
  62 |     const offlineDevice = devicesFixture[2]
  63 | 
  64 |     await expect(page.getByTestId(`device-row-${onlineDevice.id}`)).toContainText('online')
  65 |     await expect(page.getByTestId(`device-row-${offlineDevice.id}`)).toContainText('offline')
  66 |   })
  67 | })
  68 | 
```