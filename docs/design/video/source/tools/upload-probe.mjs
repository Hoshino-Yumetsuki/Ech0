// Debug: walk the real editor attach → video drop → upload → back flow, screenshot each step.
import {chromium} from 'playwright';
import {mkdir} from 'node:fs/promises';
import {serve, openFilm} from '../server.mjs';
const out = process.argv[2] || 'evidence/upload';
await mkdir(out, {recursive: true});
const {server, url} = await serve();
const browser = await chromium.launch({headless: true, executablePath: process.env.FILM_CHROMIUM || undefined});
try {
  const page = await browser.newPage({viewport: {width: 1920, height: 1080}});
  page.on('pageerror', e => console.log('pageerror', e.message.slice(0, 200)));
  page.on('console', m => { if (m.type() === 'error' || m.type() === 'warning') console.log(m.type(), m.text().slice(0, 200)); });
  await openFilm(page, url);
  await page.evaluate(() => window.__filmReady);
  await page.evaluate(() => window.seek(9.5));
  const snap = async n => { await page.waitForTimeout(400); await page.screenshot({path: `${out}/${n}.png`}); };
  const step = (n, fn) => page.evaluate(fn).then(r => { console.log(n, JSON.stringify(r)); return snap(n); });
  await step('1-attach', () => { const b = [...document.querySelectorAll('#home-layer .editor-actions__left button')]; b[1]?.click(); return b.map(x => x.getAttribute('aria-label') || x.textContent.trim()); });
  await step('2-video-tab', () => { const segs = [...document.querySelectorAll('#home-layer .seg__btn')]; segs.find(s => /video/i.test(s.textContent))?.click(); return segs.map(s => s.textContent.trim()); });
  await step('3-drop', async () => {
    const blob = await (await fetch('/demo/cats-at-dusk.mp4')).blob();
    const file = new File([blob], 'cats-at-dusk.mp4', {type: 'video/mp4'});
    const dt = new DataTransfer(); dt.items.add(file);
    const zone = document.querySelector('#home-layer .editor-media-panel button.border-dashed') || document.querySelector('#home-layer button.border-dashed');
    zone?.dispatchEvent(new DragEvent('drop', {bubbles: true, cancelable: true, dataTransfer: dt}));
    await new Promise(r => setTimeout(r, 300));
    window.__uploadFeed?.progress(0.55);
    return {zone: !!zone, feed: !!window.__uploadFeed};
  });
  await step('4-done', async () => {
    window.__uploadFeed?.finish({id: 'f-cats', key: 'videos/cats-at-dusk.mp4', url: '/demo/cats-at-dusk.mp4', content_type: 'video/mp4', size: window.__uploadFeed.total, width: 1916, height: 1080, category: 'video', storage_type: 'local'});
    await new Promise(r => setTimeout(r, 400));
    return document.querySelector('#home-layer .editor-shell')?.textContent.replace(/\s+/g, ' ').slice(0, 300);
  });
  await step('5-back', () => { document.querySelectorAll('#home-layer .editor-actions__left button')[0]?.click(); return 'back'; });
} finally { await browser.close(); server.close(); }
