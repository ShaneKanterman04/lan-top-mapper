# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: topology-filters.spec.js >> topology view >> reset focus clears selected node
- Location: tests/topology-filters.spec.js:110:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: getByText('Name Source', { exact: true })
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for getByText('Name Source', { exact: true })

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
      - generic [ref=e18]:
        - generic [ref=e19]:
          - heading "Topology Graph" [level=2] [ref=e20]
          - paragraph [ref=e21]: Gateway-centered layout derived from the latest completed scan.
        - generic [ref=e22]: 5 nodes
      - generic [ref=e23]:
        - searchbox "Search hostname, IP, vendor" [ref=e24]
        - generic [ref=e25]:
          - checkbox "Online only" [ref=e26]
          - text: Online only
        - generic [ref=e27]:
          - checkbox "Services only" [ref=e28]
          - text: Services only
        - combobox [ref=e29]:
          - option "All evidence" [selected]
          - option "arp"
          - option "icmp"
          - option "tcp"
          - option "mdns"
        - combobox [ref=e30]:
          - option "All device types" [selected]
          - option "Router or access point"
          - option "Computer"
          - option "Media device"
          - option "Printer"
        - combobox [ref=e31]:
          - option "All confidence" [selected]
          - option "High"
          - option "Medium"
          - option "Low"
        - combobox [ref=e32]:
          - option "All name sources" [selected]
          - option "dns"
          - option "ip"
          - option "mdns"
          - option "upnp-desc"
        - combobox [ref=e33]:
          - option "Concentric" [selected]
          - option "Breadth-first"
          - option "Radial"
        - button "Reset Focus" [ref=e34] [cursor=pointer]
      - generic [ref=e35]:
        - generic [ref=e36]:
          - strong [ref=e37]: Device Types
          - generic [ref=e38]:
            - generic [ref=e39]: Gateway
            - generic [ref=e41]: Router/AP
            - generic [ref=e43]: Computer
            - generic [ref=e45]: NAS
            - generic [ref=e47]: Printer
            - generic [ref=e49]: Media
            - generic [ref=e51]: Unknown
        - generic [ref=e53]:
          - strong [ref=e54]: Confidence
          - generic [ref=e55]:
            - generic [ref=e56]: High
            - generic [ref=e58]: Medium
            - generic [ref=e60]: Low
    - article [ref=e67]:
      - generic [ref=e68]:
        - generic [ref=e69]:
          - heading "Node Details" [level=2] [ref=e70]
          - paragraph [ref=e71]: Current identity, latest observation, and recent scan history.
        - button "Close" [ref=e72] [cursor=pointer]
      - paragraph [ref=e73]: Select a device to inspect its history.
```

# Test source

```ts
  12  |     const node = cy.getElementById(id)
  13  |     if (!node.length) return null
  14  |     const rendered = node.renderedPosition()
  15  |     return { x: rendered.x, y: rendered.y }
  16  |   }, nodeId)
  17  | 
  18  |   expect(position).toBeTruthy()
  19  |   await page.mouse.click(box.x + position.x, box.y + position.y)
  20  |   await page.waitForTimeout(300)
  21  | }
  22  | 
  23  | test.describe('topology view', () => {
  24  |   test.beforeEach(async ({ page }) => {
  25  |     await setupMockApi(page)
  26  |     await page.goto('/', { waitUntil: 'networkidle' })
  27  |     await page.getByTestId('tab-topology').click()
  28  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  29  |     await page.waitForFunction(() => Boolean(window.__LTM_TOPOLOGY_CY__))
  30  |   })
  31  | 
  32  |   test('renders topology graph and legend', async ({ page }) => {
  33  |     await expect(page.getByRole('heading', { name: 'Topology Graph' })).toBeVisible()
  34  |     await expect(page.getByTestId('topology-legend')).toBeVisible()
  35  |   })
  36  | 
  37  |   test('switches layout modes', async ({ page }) => {
  38  |     await page.getByTestId('filter-layout').selectOption('concentric')
  39  |     await page.waitForTimeout(400)
  40  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  41  | 
  42  |     await page.getByTestId('filter-layout').selectOption('breadthfirst')
  43  |     await page.waitForTimeout(400)
  44  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  45  | 
  46  |     await page.getByTestId('filter-layout').selectOption('radial')
  47  |     await page.waitForTimeout(400)
  48  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  49  |   })
  50  | 
  51  |   test('filters by device type', async ({ page }) => {
  52  |     await page.getByTestId('filter-device-type').selectOption('Computer')
  53  |     await page.waitForTimeout(300)
  54  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  55  |   })
  56  | 
  57  |   test('filters by confidence', async ({ page }) => {
  58  |     await page.getByTestId('filter-confidence').selectOption('high')
  59  |     await page.waitForTimeout(300)
  60  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  61  |   })
  62  | 
  63  |   test('filters by name source', async ({ page }) => {
  64  |     await page.getByTestId('filter-name-source').selectOption('dns')
  65  |     await page.waitForTimeout(300)
  66  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  67  |   })
  68  | 
  69  |   test('filters by evidence', async ({ page }) => {
  70  |     await page.getByTestId('filter-evidence').selectOption('arp')
  71  |     await page.waitForTimeout(300)
  72  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  73  |   })
  74  | 
  75  |   test('search filters topology', async ({ page }) => {
  76  |     await page.getByTestId('topology-search').fill('Printer')
  77  |     await page.waitForTimeout(300)
  78  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  79  |   })
  80  | 
  81  |   test('online only toggle filters topology', async ({ page }) => {
  82  |     await page.getByTestId('filter-online-only').locator('input').check()
  83  |     await page.waitForTimeout(300)
  84  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  85  |   })
  86  | 
  87  |   test('services only toggle filters topology', async ({ page }) => {
  88  |     await page.getByTestId('filter-services-only').locator('input').check()
  89  |     await page.waitForTimeout(300)
  90  |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  91  |   })
  92  | 
  93  |   test('selects gateway and updates detail panel', async ({ page }) => {
  94  |     await clickNode(page, 'gateway')
  95  |     await expect(page.getByText('Name Source', { exact: true })).toBeVisible()
  96  |     await expect(page.getByRole('heading', { name: 'Identity Claims' })).toBeVisible()
  97  |   })
  98  | 
  99  |   test('selects a non-gateway node and updates detail panel', async ({ page }) => {
  100 |     const nodeId = await page.evaluate(() => {
  101 |       const cy = window.__LTM_TOPOLOGY_CY__
  102 |       const node = cy.nodes().filter((item) => item.data('kind') !== 'gateway').first()
  103 |       return node?.id()
  104 |     })
  105 |     expect(nodeId).toBeTruthy()
  106 |     await clickNode(page, nodeId)
  107 |     await expect(page.getByText('Recent Observations')).toBeVisible()
  108 |   })
  109 | 
  110 |   test('reset focus clears selected node', async ({ page }) => {
  111 |     await clickNode(page, 'gateway')
> 112 |     await expect(page.getByText('Name Source', { exact: true })).toBeVisible()
      |                                                                  ^ Error: expect(locator).toBeVisible() failed
  113 | 
  114 |     await page.getByTestId('reset-focus').click()
  115 |     await page.waitForTimeout(300)
  116 |     await expect(page.getByTestId('device-detail-empty')).toBeVisible()
  117 |   })
  118 | 
  119 |   test('renders empty topology gracefully', async ({ page }) => {
  120 |     await setupMockApi(page, { topology: emptyTopologyFixture })
  121 |     await page.reload()
  122 |     await page.getByTestId('tab-topology').click()
  123 |     await expect(page.getByTestId('topology-canvas')).toBeVisible()
  124 |   })
  125 | })
  126 | 
```