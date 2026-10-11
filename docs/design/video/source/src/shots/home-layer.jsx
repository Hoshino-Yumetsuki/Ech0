// The real Ech0 home view (HomeView → HomePage: editor, timeline, Status widgets), mounted once
// through the Vue bridge and filmed through a camera wrapper. Product state is driven by real
// interactions (typing, tag picker, Publish, sidebar Status link) and is a function of film time.
import React from 'react';
import {VueHost} from '../kit/vue-host.jsx';
import {Cursor, paintCursor} from '../kit/cursor.jsx';
import {setFieldValue} from '../kit/interact.js';
import {shot, onDrive, onRender, frame} from '../engine.js';
import {setPublished, isPublished} from '../fixtures/api.js';
import {Rings, pulseRings, ringRadius, BEAT, $} from './common.jsx';
import {DOT} from './rail.jsx';
import {T as DT} from './detail.jsx';

// The demo upload: a 5 s clip from the owner's own instance (data/files/videos), served as /demo/cats-at-dusk.mp4.
export const UPLOAD = {name: 'cats-at-dusk.mp4', url: '/demo/cats-at-dusk.mp4', width: 1916, height: 1080, duration: 5.04};
let demoFile = null;
const loadDemoFile = async () => { if (!demoFile) { const blob = await (await fetch(UPLOAD.url)).blob(); demoFile = new File([blob], UPLOAD.name, {type: 'video/mp4'}); } return demoFile; };
export const POST = 'Moved my notes off the feed and onto my own server tonight.\n\nSame words. **My address.**';
const S = id => shot(id);
const T = {};   // absolute action times, filled in buildHome from plan.json shot starts
const MARKS = [
  // Each destination is lit by its own ring (index into the three reach rings), so they light in turn.
  {id: 'connect', ring: 1, label: 'Connected Ech0 sites', sel: '.home-status-widgets > :nth-child(3) .widget'},
  {id: 'comment', ring: 2, label: 'Comments', sel: '.home-status-widgets > :nth-child(4) .widget'},
  {id: 'rss', ring: 0, label: 'RSS feed', sel: '.home-header a[aria-label="RSS"]'},
];

export function HomeLayer() {
  return <div id="home-layer">
    <div className="home-cam"><VueHost name="home" className="home-host" /></div>
    <Rings name="land" count={2} className="land-rings" />
    <div className="hl-marks">{MARKS.map(m => <div key={m.id} className="hl-mark" data-mark={m.id}><i /><span>{m.label}</span></div>)}</div>
    <div className="drag-file"><i /><b>cats-at-dusk.mp4</b><span>1.2 MB</span></div>
    <Cursor id="home" color="#33251d" stroke="#fffaf2" />
  </div>;
}

const B = () => window.Ech0Bridge;
const main = () => $('#home-layer .home-main');
const center = el => { const r = el.getBoundingClientRect(); return {x: r.left + r.width / 2, y: r.top + r.height / 2, r}; };
const until = async (cond, ms = 1500) => { const t0 = performance.now(); while (!cond() && performance.now() - t0 < ms) await new Promise(r => setTimeout(r, 16)); await frame(); };
const easeInOut = u => (u < 0.5 ? 2 * u * u : 1 - Math.pow(-2 * u + 2, 2) / 2);
const clamp01 = u => Math.max(0, Math.min(1, u));

function phaseAt(t) {
  if (t < T.publish) return 'write';
  if (DT.open !== undefined && t >= DT.open && t < DT.back + 0.4) return 'away';   // detail.jsx owns the route
  if (t < T.status) return 'timeline';
  if (t < T.back) return 'status';
  return 'timeline';
}
function scrollAt(t) {
  const st = S('stream');
  if (t < st.start + 0.3) return 0;
  if (t < st.end - 0.2) return T.scrollMax * easeInOut(clamp01((t - st.start - 0.3) / (st.end - st.start - 0.5)));
  if (t < T.reachBack) return T.scrollMax;
  if (t < T.reachBack + 1.0) return T.scrollMax * (1 - easeInOut(clamp01(t - T.reachBack)));
  return 0;
}

let lastT = -1, busy = null;
async function driveHome(t) {
  if (t === lastT) return busy;
  const back = t < lastT;
  lastT = t;
  busy = (async () => {
    const b = B();
    const phase = phaseAt(t);
    const route = b.route();

    if (phase === 'away') return;
    if (phase === 'write') {
      if (!route.includes('tab=publish') || back) {
        setPublished(false);
        if (back) await b.resetEditor();
        await b.navigate('/?tab=publish');
        await until(() => $('#home-layer .editor-shell textarea'));
      }
      if (back) window.__uploadFeed = null;
      // Attachment: real "Add attachment" → Media panel → Video tab → drop the file → upload → back.
      const leftBtns = () => [...document.querySelectorAll('#home-layer .editor-actions__left button')];
      const mediaOpen = () => !!$('#home-layer .editor-media-panel');
      const wantMedia = t >= T.attach && t < T.toText;
      if (mediaOpen() !== wantMedia) {
        (wantMedia ? leftBtns()[1] : leftBtns()[0])?.click();
        await until(() => mediaOpen() === wantMedia, 800);
      }
      if (mediaOpen()) {
        const seg = [...document.querySelectorAll('#home-layer .editor-media-panel .seg__btn')].find(x => /video/i.test(x.textContent));
        if (t >= T.video && seg && !seg.classList.contains('seg__btn--active')) { seg.click(); await frame(); await frame(); }
        const zone = $('#home-layer .editor-media-panel button.border-dashed');
        if (t >= T.drop && !window.__uploadFeed && zone) {
          const dt = new DataTransfer(); dt.items.add(await loadDemoFile());
          zone.dispatchEvent(new DragEvent('drop', {bubbles: true, cancelable: true, dataTransfer: dt}));
          await until(() => !!window.__uploadFeed, 800);
        }
      }
      const feed = window.__uploadFeed;
      if (feed && !feed.done) {
        feed.progress(Math.max(0, Math.min(1, (t - T.drop) / (T.upDone - T.drop - 0.12))));
        if (t >= T.upDone) feed.finish({id: 'f-cats', key: 'videos/' + UPLOAD.name, url: UPLOAD.url, content_type: 'video/mp4', size: feed.total, width: UPLOAD.width, height: UPLOAD.height, category: 'video', storage_type: 'local'});
        await new Promise(r => setTimeout(r, 30)); await frame();
      }
      const field = $('#home-layer .editor-shell textarea');
      const n = t < T.typeAt ? 0 : Math.min(POST.length, Math.floor((t - T.typeAt) * T.cps) + 1);
      if (field) setFieldValue(field, POST.slice(0, n));
      // Tag picker: real Popover button + real chip toggle.
      const btn = $('#home-layer .editor-actions__tag button');
      const panel = () => $('#home-layer .editor-actions__tag-panel');
      const wantOpen = t >= T.tagOpen && t < T.tagClose;
      if (!!panel() !== wantOpen && btn) { btn.click(); await until(() => !!panel() === wantOpen, 500); }
      const chip = panel() && [...panel().querySelectorAll('.editor-actions__tag-chip')].find(c => c.textContent.trim() === '#notes');
      if (chip) {
        const on = chip.classList.contains('editor-actions__tag-chip--selected');
        if (on !== t >= T.tagPick) { chip.click(); await frame(); }
      }
      // Publish trigger opens the real public/private choice.
      const pubBtn = $('#home-layer .editor-actions__publish button');
      const pubPanel = () => $('#home-layer .editor-actions__publish-panel');
      const wantPub = t >= T.pubOpen;
      if (!!pubPanel() !== wantPub && pubBtn) { pubBtn.click(); await until(() => !!pubPanel() === wantPub, 500); }
    } else {
      if (route.includes('tab=publish')) {
        const publish = $('#home-layer .editor-actions__publish-option');   // "Publish public"
        const field = $('#home-layer .editor-shell textarea');
        if (!back && field && field.value === POST && publish) publish.click();   // the real Publish path
        else { setPublished(true); await b.navigate('/'); await b.refreshEchos(); }
        await until(() => !b.route().includes('tab=publish') && $('#home-layer .timeline-marker'), 2500);
        await until(() => $('#home-layer .home-main')?.textContent.includes('My address'), 1500);
        await new Promise(r => setTimeout(r, 120)); await frame();
      }
      if (!isPublished()) { setPublished(true); await b.refreshEchos(); await until(() => $('#home-layer .home-main')?.textContent.includes('My address'), 1500); await new Promise(r => setTimeout(r, 120)); await frame(); }   // jumped past Publish (cards reveal via IntersectionObserver)
      const want = phase === 'status' ? '/?tab=status' : '/';
      if (b.route() !== want) {
        const link = [...document.querySelectorAll('#home-layer .home-aside--rail a')].find(a => a.textContent.trim() === 'Status');
        if (want === '/?tab=status' && link && !back) link.click();
        else await b.navigate(want);
        await until(() => b.route() === want && (want !== '/?tab=status' || $('#home-layer .home-status-widgets > :nth-child(4)')), 2500);
        if (want === '/?tab=status') await new Promise(r => setTimeout(r, 250));   // widget fixtures settle
      }
      const m = main();
      if (m && phase === 'timeline') { const y = Math.round(scrollAt(t)); if (m.scrollTop !== y) m.scrollTop = y; }
    }
  })();
  return busy;
}

// Film cursor: waypoints are live product elements (measured each frame, screen space).
function cursorPath(t) {
  const tag = $('#home-layer .editor-actions__tag button');
  const chip = [...document.querySelectorAll('#home-layer .editor-actions__tag-chip')].find(c => c.textContent.trim() === '#notes');
  const pub = $('#home-layer .editor-actions__publish button');
  const pubOpt = $('#home-layer .editor-actions__publish-option');
  const status = [...document.querySelectorAll('#home-layer .home-aside--rail a')].find(a => a.textContent.trim() === 'Status');
  const P = el => (el ? center(el) : null);
  const attach = [...document.querySelectorAll('#home-layer .editor-actions__left button')][1];
  const backBtn = [...document.querySelectorAll('#home-layer .editor-actions__left button')][0];
  const videoSeg = [...document.querySelectorAll('#home-layer .editor-media-panel .seg__btn')].find(x => /video/i.test(x.textContent));
  const zone = $('#home-layer .editor-media-panel button.border-dashed');
  if (t >= T.attach - 0.8 && t < T.publish + 0.6) {
    const pts = [
      {at: T.attach - 0.8, p: {x: 1720, y: 930}},
      {at: T.attach - 0.12, p: P(attach)},
      {at: T.video - 0.12, p: P(videoSeg) || P(attach)},
      {at: T.dragFrom, p: {x: 1905, y: 800}},          // off to the right: pick the file up
      {at: T.drop - 0.05, p: P(zone) || {x: 1330, y: 700}},
      {at: T.toText - 0.12, p: P(backBtn)},
      {at: T.tagOpen - 0.1, p: P(tag)},
      {at: T.tagPick - 0.15, p: P(chip) || P(tag)},
      {at: T.tagClose - 0.1, p: P(tag)},
      {at: T.pubOpen - 0.12, p: P(pub)},
      {at: T.publish - 0.12, p: P(pubOpt) || P(pub)},
    ].filter(k => k.p);
    return {keys: pts.map(k => ({at: k.at, x: k.p.x, y: k.p.y})), clicks: [T.attach, T.video, T.drop, T.toText, T.tagOpen, T.tagPick, T.tagClose, T.pubOpen, T.publish], show: [T.attach - 0.8, T.publish + 0.3]};
  }
  if (t >= T.status - 0.9 && t < T.status + 0.7 && status) {
    const p = center(status);
    return {keys: [{at: T.status - 0.9, x: 1740, y: 980}, {at: T.status - 0.12, x: p.x, y: p.y}], clicks: [T.status], show: [T.status - 0.9, T.status + 0.7]};
  }
  return null;
}

export function buildHome(tl) {
  const w = S('write'), land = S('land'), reach = S('reach'), cop = S('copilot');
  Object.assign(T, {
    // All clicks land on beats (96 BPM): see plan.json actions.
    attach: w.start + 1.25, video: w.start + 1.875, dragFrom: w.start + 2.5, drop: w.start + 3.75, upDone: w.start + 5.75, toText: w.start + 6.25,
    typeAt: w.start + 6.875, cps: 25,
    tagOpen: w.start + 10.625, tagPick: w.start + 11.25, tagClose: w.start + 11.875,
    pubOpen: w.start + 12.5, publish: w.start + 13.125,
    reachBack: reach.start, status: reach.start + 1.25, back: cop.start + 1.0,
    scrollMax: 1180,
  });
  const layer = $('#home-layer'), cam = $('#home-layer .home-cam');
  // Visibility: in with the editor, out under Copilot (the admin panel takes over after that).
  tl.set(layer, {opacity: 0}, 0);
  tl.fromTo(layer, {opacity: 0, y: 70}, {opacity: 1, y: 0, duration: 0.9, ease: 'expo.out', immediateRender: false}, w.start - 0.1);
  tl.to(layer, {opacity: 0, filter: 'blur(8px)', duration: 0.5, ease: 'power2.in'}, cop.start - 0.2);
  // Camera (transform-origin 0 0): page point P → screen point Q at scale s means x = Qx - s·Px.
  const frameAt = (s, px, py, qx, qy) => ({scale: s, x: qx - s * px, y: qy - s * py});
  tl.set(cam, {transformOrigin: '0 0', ...frameAt(1.85, 825, 190, 1330, 450)}, 0);
  tl.to(cam, {...frameAt(1.6, 825, 300, 1330, 560), duration: 0.9, ease: 'power2.inOut'}, T.toText + 0.1);
  tl.to(cam, {...frameAt(1.85, 825, 300, 1330, 520), duration: 1.0, ease: 'power2.inOut'}, T.publish + 0.3);
  // Land: a slow push toward the new Echo while the caption is read.
  tl.to(cam, {...frameAt(1.98, 825, 300, 1330, 520), duration: land.end - land.start - 1.2, ease: 'sine.inOut'}, land.start + 1.0);
  tl.to(cam, {...frameAt(1.2, 960, 400, 1330, 560), duration: 1.0, ease: 'power2.inOut'}, reach.start);
  tl.to(cam, {...frameAt(1.3, 826, 450, 1330, 570), duration: 1.0, ease: 'power2.inOut'}, T.status + 0.4);
  // After the rings land, push in until the Connect and Comment cards read at video scale.
  tl.to(cam, {...frameAt(1.85, 826, 690, 1330, 610), duration: 1.3, ease: 'power2.inOut'}, reach.start + 5.35);

  for (const id of ['write', 'land', 'stream', 'reach']) onDrive(id, local => driveHome(local + S(id).start));
  // The file travels with the pointer from the frame edge into the drop zone.
  const chip = $('#home-layer .drag-file');
  onRender('write', (local, t) => {
    const on = t >= T.dragFrom && t < T.drop + 0.12;
    const cur = $('#home-layer [data-cursor="home"]');
    chip.style.opacity = on ? String(Math.min(1, (t - T.dragFrom) / 0.15, (T.drop + 0.12 - t) / 0.12)) : '0';
    if (on && cur) chip.style.transform = cur.style.transform.replace(/translate\(([-\d.]+)px, ([-\d.]+)px\)/, (m, x, y) => `translate(${+x + 22}px, ${+y + 26}px)`);
  });
  // Uploaded clip: every <video> of it shows the frame for film time (muted, looping), seek-exact.
  const syncVideos = async t => {
    const vids = [...document.querySelectorAll('#home-layer video, #detail-layer video')].filter(v => (v.currentSrc || v.src || '').includes('cats-at-dusk'));
    await Promise.all(vids.map(v => {
      v.muted = true; v.pause();
      // Editor preview runs from the finished upload; the published Echo restarts it from Publish.
      const want = Math.max(0, t < T.publish ? t - T.upDone : t - T.publish) % (UPLOAD.duration - 0.05);
      if (v.readyState >= 1 && Math.abs(v.currentTime - want) < 0.02) return null;
      return new Promise(res => { const done = () => res(); v.addEventListener('seeked', done, {once: true}); setTimeout(done, 600); v.currentTime = want; });
    }));
  };
  for (const id of ['write', 'land', 'detail', 'stream']) onDrive(id, local => syncVideos(local + S(id).start));
  for (const id of ['write', 'reach']) onRender(id, (local, t) => {
    const c = cursorPath(t);
    paintCursor($('#home-layer [data-cursor="home"]'), t, c?.keys || [], c?.clicks || [], c?.show || [-1, -1]);
  });

  // Land: the new Echo's real date dot rings.
  const landRings = $('#home-layer [data-rings="land"]');
  pulseRings(tl, landRings, land.start + BEAT, {count: 2, maxR: 460, dur: 1.8, gap: BEAT});
  onRender('land', () => {
    const dot = $('#home-layer .timeline-marker');
    if (dot) { const c = center(dot); landRings.style.transform = `translate(${c.x}px, ${c.y}px)`; }
  });

  // Reach: rings from the film dot sweep across the real Status tab; each widget they touch lights up.
  const ringStarts = [0, 1, 2].map(i => reach.start + 2.5 + i * BEAT);
  pulseRings(tl, $('#rail [data-rings="dot"]'), ringStarts[0], {count: 3, maxR: 1700, dur: 2.6, from: 0.75});
  tl.set('#home-layer .hl-marks', {opacity: 0}, 0);
  tl.set('#home-layer .hl-marks', {opacity: 1}, reach.start);
  tl.to('#home-layer .hl-marks', {opacity: 0, duration: 0.3}, reach.end - 0.4);
  onRender('reach', (local, t) => {
    for (const m of MARKS) {
      const box = $(`#home-layer [data-mark="${m.id}"]`), el = $('#home-layer ' + m.sel);
      if (!el || t < T.status + 0.2) { box.style.opacity = '0'; continue; }
      const c = center(el);
      const d = Math.hypot(c.x - DOT.rx, c.y - DOT.ry);
      const s0 = ringStarts[m.ring];
      let hit = Infinity;
      for (let tt = s0; tt < s0 + 2.6; tt += 1 / 120) if (ringRadius(tt, s0, 1700, 2.6) >= d) { hit = tt; break; }
      (window.__ringHits ??= {})[m.id] = Math.round(hit * 1000) / 1000;   // for SFX cue placement (plan.json)
      const since = t - hit;
      const pad = m.id === 'rss' ? 10 : 14;
      box.style.opacity = since >= 0 ? String(Math.min(1, since / 0.25)) : '0';
      box.classList.toggle('hl-mark--below', c.r.top < 90);
      box.style.transform = `translate(${c.r.left - pad}px, ${c.r.top - pad}px)`;
      box.style.width = c.r.width + pad * 2 + 'px';
      box.style.height = c.r.height + pad * 2 + 'px';
      box.querySelector('i').style.transform = since >= 0 ? `scale(${1 + 0.06 * Math.max(0, 1 - since / 0.4)})` : '';
    }
  });
}
