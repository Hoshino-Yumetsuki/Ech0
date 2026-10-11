// Debug helper: node tools/probe.mjs <seconds> "<js expression>"
import {chromium} from 'playwright';
import {serve, openFilm} from '../server.mjs';
const [t = '1', expr = 'document.title', shotPath] = process.argv.slice(2);
const {server, url} = await serve();
const browser = await chromium.launch({headless: true, executablePath: process.env.FILM_CHROMIUM || undefined});
try {
  const page = await browser.newPage({viewport: {width: 1920, height: 1080}});
  page.on('console', m => { if (m.type() === 'error') console.log('console.error', m.text().slice(0, 300)); });
  page.on('pageerror', e => console.log('pageerror', e.message));
  await openFilm(page, url);
  await page.evaluate(() => window.__filmReady);
  await page.evaluate(x => window.seek(x), Number(t));
  console.log(JSON.stringify(await page.evaluate(expr), null, 1));
  if (shotPath) { await page.waitForTimeout(600); await page.screenshot({path: shotPath}); }
} finally { await browser.close(); server.close(); }
