// Debug: mount the real Panel over the film, visit each route, screenshot + log API calls.
// node tools/panel-tour.mjs <outDir> [/panel/dashboard ...]   (TOUR_SHOT=echo to mount EchoView instead)
import {chromium} from 'playwright';
import {mkdir} from 'node:fs/promises';
import {serve, openFilm} from '../server.mjs';
const [out = 'evidence/panel', ...routes] = process.argv.slice(2);
const list = routes.length ? routes : ['/panel/dashboard', '/panel/setting', '/panel/user', '/panel/storage', '/panel/data-management', '/panel/comment', '/panel/sso', '/panel/extension', '/panel/advance', '/panel/system-log'];
await mkdir(out, {recursive: true});
const {server, url} = await serve();
const browser = await chromium.launch({headless: true, executablePath: process.env.FILM_CHROMIUM || undefined});
try {
  const page = await browser.newPage({viewport: {width: 1920, height: 1080}});
  page.on('pageerror', e => console.log('pageerror', e.message.slice(0, 200)));
  await openFilm(page, url);
  await page.evaluate(() => window.__filmReady);
  const first = list[0].split('|')[0];
  if (process.env.TOUR_SHOT) await page.evaluate(r => window.Ech0Bridge.navigate(r), first);   // views that read route params on setup
  await page.evaluate(async shotName => {
    const el = document.createElement('div');
    el.id = 'panel-debug';
    el.style.cssText = 'position:fixed;inset:0;z-index:9999;overflow:auto;background:var(--color-bg-canvas)';
    document.body.appendChild(el);
    await window.Ech0Bridge.mount(el, shotName);
  }, process.env.TOUR_SHOT || 'panel');
  for (const spec of list) {
    const [r, tab] = spec.split('|');   // e.g. /panel/comment|Comment Management clicks that segmented tab
    const before = await page.evaluate(() => window.__filmApiLog.length);
    await page.evaluate(r => window.Ech0Bridge.navigate(r), r);
    await page.waitForTimeout(900);
    if (tab) await page.evaluate(tab => [...document.querySelectorAll('#panel-debug button')].find(b => b.textContent.trim() === tab)?.click(), tab);
    await page.waitForTimeout(900);
    const api = await page.evaluate(n => window.__filmApiLog.slice(n), before);
    console.log(spec, JSON.stringify([...new Set(api)]));
    await page.screenshot({path: `${out}/${(r + (tab ? '-' + tab : '')).replaceAll('/', '_').replaceAll(' ', '-')}.png`});
  }
} finally { await browser.close(); server.close(); }
