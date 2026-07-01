import { chromium, request } from 'playwright'
import { mkdir } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '..')
const outDir = path.join(root, 'docs', 'screenshots')
const baseURL = process.env.SCREENSHOT_BASE_URL || 'http://localhost:5173'
const apiURL = process.env.SCREENSHOT_API_URL || 'http://localhost:8080'
const email = process.env.SCREENSHOT_EMAIL || 'screenshot@switchboard.local'
const password = process.env.SCREENSHOT_PASSWORD || 'screenshot-demo'

const shots = [
  { name: 'deployment-reports', path: '/security/reports', wait: 1000 },
  { name: 'cve-dashboard', path: '/security/cves', wait: 1000 },
  { name: 'profile-notifications', path: '/profile', wait: 600 },
]

await mkdir(outDir, { recursive: true })

const api = await request.newContext({ baseURL: apiURL })
const login = await api.post('/api/auth/login', {
  data: { email, password, remember_me: false },
})
if (!login.ok()) {
  throw new Error(`Login failed (${login.status()}): ${await login.text()}`)
}
const storage = await api.storageState()
await api.dispose()

const browser = await chromium.launch()
const context = await browser.newContext({
  viewport: { width: 1440, height: 900 },
  deviceScaleFactor: 2,
  storageState: storage,
})
const page = await context.newPage()

for (const shot of shots) {
  await page.goto(`${baseURL}${shot.path}`, { waitUntil: 'networkidle' })
  await page.waitForSelector('.page-header, .dashboard-hero, h1', { timeout: 15000 })
  await page.waitForTimeout(shot.wait)
  await page.screenshot({
    path: path.join(outDir, `${shot.name}.png`),
    fullPage: false,
  })
  console.log(`saved ${shot.name}.png`)
}

await browser.close()
console.log(`Screenshots written to ${outDir}`)
