# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: scan-flow.spec.js >> scan flow >> scan fails and updates status
- Location: tests/scan-flow.spec.js:72:3

# Error details

```
Error: expect(locator).toHaveText(expected) failed

Locator:  getByTestId('status-pill')
Expected: "failed"
Received: "idle"
Timeout:  5000ms

Call log:
  - Expect "toHaveText" with timeout 5000ms
  - waiting for getByTestId('status-pill')
    9 × locator resolved to <span class="status-pill" data-testid="status-pill">idle</span>
      - unexpected value "idle"

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
  3  | import {
  4  |   summaryFixture,
  5  |   scanStartFixture,
  6  |   scanStatusRunningFixture,
  7  |   scanStatusCompletedFixture,
  8  |   scanStatusFailedFixture,
  9  | } from './fixtures/api-responses.js'
  10 | 
  11 | test.describe('scan flow', () => {
  12 |   test.beforeEach(async ({ page }) => {
  13 |     await page.goto('/', { waitUntil: 'networkidle' })
  14 |   })
  15 | 
  16 |   test('start scan button submits scan and shows running state', async ({ page }) => {
  17 |     await setupMockApi(page, {
  18 |       summary: { ...summaryFixture, latestScanState: 'idle' },
  19 |       scanStart: scanStartFixture,
  20 |       scanStatus: scanStatusRunningFixture,
  21 |     })
  22 |     await page.reload()
  23 | 
  24 |     await expect(page.getByTestId('start-scan-button')).toBeEnabled()
  25 |     await expect(page.getByTestId('status-pill')).toHaveText('idle')
  26 | 
  27 |     await page.getByTestId('start-scan-button').click()
  28 | 
  29 |     await expect(page.getByTestId('start-scan-button')).toBeDisabled()
  30 |     await expect(page.getByTestId('start-scan-button')).toHaveText('Scanning...')
  31 |     await expect(page.getByTestId('status-pill')).toHaveText('running')
  32 |   })
  33 | 
  34 |   test('scan completes and updates status', async ({ page }) => {
  35 |     let pollCount = 0
  36 |     await setupMockApi(page, {
  37 |       summary: { ...summaryFixture, latestScanState: 'idle' },
  38 |       scanStart: scanStartFixture,
  39 |       scanStatus: scanStatusCompletedFixture,
  40 |     })
  41 | 
  42 |     await page.route(/\/api\/scans\/\d+/, async (route) => {
  43 |       pollCount++
  44 |       const status = pollCount <= 2 ? scanStatusRunningFixture : scanStatusCompletedFixture
  45 |       return route.fulfill({
  46 |         status: 200,
  47 |         contentType: 'application/json',
  48 |         body: JSON.stringify(status),
  49 |       })
  50 |     })
  51 | 
  52 |     await page.reload()
  53 |     await page.getByTestId('start-scan-button').click()
  54 | 
  55 |     await page.waitForTimeout(4000)
  56 |     await expect(page.getByTestId('status-pill')).toHaveText('idle')
  57 |   })
  58 | 
  59 |   test('scan start failure shows error', async ({ page }) => {
  60 |     await setupMockApi(page, {
  61 |       summary: { ...summaryFixture, latestScanState: 'idle' },
  62 |       scanStart: null,
  63 |     })
  64 |     await page.reload()
  65 | 
  66 |     await page.getByTestId('start-scan-button').click()
  67 | 
  68 |     await expect(page.getByTestId('error-box')).toBeVisible()
  69 |     await expect(page.getByTestId('error-box')).toContainText('Scan start failed')
  70 |   })
  71 | 
  72 |   test('scan fails and updates status', async ({ page }) => {
  73 |     await setupMockApi(page, {
  74 |       summary: { ...summaryFixture, latestScanState: 'idle' },
  75 |       scanStart: scanStartFixture,
  76 |       scanStatus: scanStatusFailedFixture,
  77 |     })
  78 |     await page.reload()
  79 | 
  80 |     await page.getByTestId('start-scan-button').click()
  81 |     await page.waitForTimeout(2000)
  82 | 
> 83 |     await expect(page.getByTestId('status-pill')).toHaveText('failed')
     |                                                   ^ Error: expect(locator).toHaveText(expected) failed
  84 |   })
  85 | 
  86 |   test('button disabled when scan already running', async ({ page }) => {
  87 |     await setupMockApi(page, {
  88 |       summary: { ...summaryFixture, latestScanState: 'running' },
  89 |     })
  90 |     await page.reload()
  91 | 
  92 |     await expect(page.getByTestId('start-scan-button')).toBeDisabled()
  93 |     await expect(page.getByTestId('start-scan-button')).toHaveText('Scanning...')
  94 |     await expect(page.getByTestId('status-pill')).toHaveText('running')
  95 |   })
  96 | })
  97 | 
```