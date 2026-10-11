// The real Ech0 admin panel (PanelView → PanelPage + child pages), mounted once through the Vue
// bridge and filmed through a camera wrapper. Pages are reached by real clicks on the panel nav,
// the pages' segmented tabs and the file tree; the system log receives live lines by film time.
import React from 'react';
import {VueHost} from '../kit/vue-host.jsx';
import {Cursor, paintCursor} from '../kit/cursor.jsx';
import {offsetCenter} from '../kit/measure.js';
import {shot, onDrive, onRender, frame} from '../engine.js';
import {LIVE_LOGS, logEntry} from '../fixtures/panel.js';
import {$} from './common.jsx';

export function PanelLayer() {
  return <div id="panel-layer">
    <div className="panel-cam"><VueHost name="panel" className="panel-host" /></div>
    <Cursor id="panel" color="#33251d" stroke="#fffaf2" />
  </div>;
}

const B = () => window.Ech0Bridge;
const q = sel => $('#panel-layer ' + sel);
const qa = sel => [...document.querySelectorAll('#panel-layer ' + sel)];
const byText = (sel, text) => qa(sel).find(el => el.textContent.trim() === text);
const center = el => { const r = el.getBoundingClientRect(); return {x: r.left + r.width / 2, y: r.top + r.height / 2}; };
const until = async (cond, ms = 1500) => { const t0 = performance.now(); while (!cond() && performance.now() - t0 < ms) await new Promise(r => setTimeout(r, 16)); await frame(); };

let STEPS = [];
function stepAt(t) { let s = STEPS[0]; for (const x of STEPS) if (t >= x.at) s = x; return s; }

let lastT = -1, busy = null;
async function drivePanel(t) {
  if (t === lastT) return busy;
  const back = t < lastT;
  lastT = t;
  busy = (async () => {
    const b = B(), step = stepAt(t);
    if (b.route() !== step.route) {
      const nav = step.nav && byText('.panel-nav button', step.nav);
      if (nav && !back && t - step.at < 0.5) nav.click(); else await b.navigate(step.route);
      await until(() => b.route() === step.route && q('.panel-main')?.textContent.length > 0);
      await new Promise(r => setTimeout(r, 200));   // page fixtures settle
      await frame();
    }
    if (step.tab) {
      const [label, at] = step.tab;
      const want = t >= at ? label : qa('.seg__btn')[0]?.textContent.trim();
      const btn = byText('.seg__btn', want);
      if (btn && !btn.classList.contains('seg__btn--active')) { btn.click(); await new Promise(r => setTimeout(r, 250)); await frame(); }
    }
    if (step.tree) {
      const [rootAt, folder, folderAt] = step.tree;
      const root = q('.tree-row.root-row');
      const rootOpen = () => q('.tree-row.root-row .node-icon')?.textContent.trim() === '▾';
      if (root && rootOpen() !== t >= rootAt) { root.click(); await until(() => rootOpen() === t >= rootAt, 800); await new Promise(r => setTimeout(r, 120)); }
      const row = qa('.tree-row').find(r => r.querySelector('.node-name')?.textContent.trim() === folder);
      const opened = () => qa('.tree-row .node-name').some(n => n.textContent.trim().endsWith('.jpg'));
      if (row && opened() !== t >= folderAt) { row.click(); await until(() => opened() === t >= folderAt, 800); await new Promise(r => setTimeout(r, 120)); }
    }
    if (step.format) {
      // Export format cards (Snapshot / Capsule): selecting one only changes the chosen format.
      const [label, at] = step.format;
      const card = name => qa('.export-format-card').find(c => c.querySelector('h3')?.textContent.trim() === name);
      const want = card(t >= at ? label : 'Snapshot');
      if (want && !want.classList.contains('active')) { want.click(); await frame(); await frame(); }
    }
    if (step.logs) {
      const feed = window.__logFeed;
      if (feed && feed.readyState === 1) {
        feed.__n ??= 0;
        let pushed = false;
        while (feed.__n < LIVE_LOGS.length && step.at + LIVE_LOGS[feed.__n][0] <= t) {
          const [at, level, module, msg, fields] = LIVE_LOGS[feed.__n++];
          feed.push(logEntry(at, level, module, msg, fields));
          pushed = true;
        }
        if (pushed) { await new Promise(r => setTimeout(r, 30)); await frame(); }
      }
    }
  })();
  return busy;
}

// Cursor: moves to each click target shortly before its click (targets measured live).
function cursorKeys(t) {
  const keys = [{at: CLICKS[0].at - 0.9, p: {x: 1780, y: 1000}}];
  for (const c of CLICKS) {
    const el = c.el();
    const p = el ? center(el) : keys[keys.length - 1].p;
    keys.push({at: c.at - 0.12, p});
  }
  return keys.map(k => ({at: k.at, x: k.p.x, y: k.p.y}));
}
let CLICKS = [];

export function buildPanel(tl) {
  const p = shot('panel'), own = shot('own'), close = shot('close');
  const A = rel => p.start + rel;
  STEPS = [
    {at: p.start - 1, route: '/panel/dashboard'},
    {at: A(5.0), route: '/panel/comment', nav: 'Comments', tab: ['Comment Management', A(5.625)]},
    {at: A(7.5), route: '/panel/storage', nav: 'Storage', tab: ['File Manager', A(8.125)], tree: [A(8.75), 'images', A(9.375)]},
    {at: A(10.0), route: '/panel/extension', nav: 'Extensions', tab: ['MCP', A(10.625)]},
    {at: A(13.75), route: '/panel/system-log', nav: 'Logs', logs: true},
    {at: own.start, route: '/panel/data-management', nav: 'Data', tab: ['Export', own.start + 0.625], format: ['Capsule', own.start + 5.0]},
  ];
  CLICKS = [];
  for (const s of STEPS.slice(1)) {
    CLICKS.push({at: s.at, el: () => byText('.panel-nav button', s.nav)});
    if (s.tab) CLICKS.push({at: s.tab[1], el: () => byText('.seg__btn', s.tab[0])});
    if (s.tree) {
      CLICKS.push({at: s.tree[0], el: () => q('.tree-row.root-row')});
      CLICKS.push({at: s.tree[2], el: () => qa('.tree-row').find(r => r.querySelector('.node-name')?.textContent.trim() === s.tree[1])});
    }
    if (s.format) CLICKS.push({at: s.format[1], el: () => qa('.export-format-card').find(c => c.querySelector('h3')?.textContent.trim() === s.format[0])});
  }

  const layer = $('#panel-layer'), cam = $('#panel-layer .panel-cam'), host = $('#panel-layer .panel-host');
  tl.set(layer, {opacity: 0}, 0);
  tl.fromTo(layer, {opacity: 0, y: 70}, {opacity: 1, y: 0, duration: 0.9, ease: 'expo.out', immediateRender: false}, p.start - 0.05);
  tl.set(layer, {opacity: 0}, close.start);

  // Camera. Geometry comes from the real layout (offset chain, unaffected by transforms).
  const shell = q('.panel-shell'), nav = q('.panel-nav');
  const sc = offsetCenter(shell, host), navTop = offsetCenter(nav, host).y - nav.offsetHeight / 2;
  const shellLeft = sc.x - shell.offsetWidth / 2;
  const frameAt = (s, px, py, qx, qy) => ({scale: s, x: qx - s * px, y: qy - s * py});
  const full = frameAt(1.5, sc.x, navTop, 960, 262);
  const navRight = offsetCenter(nav, host).x + nav.offsetWidth / 2;
  tl.set(cam, {transformOrigin: '0 0', ...frameAt(0.86, sc.x, sc.y - 40, 1350, 560)}, 0);
  tl.to(cam, {...full, duration: 1.2, ease: 'power3.inOut'}, A(2.5));
  // Soft masks (framing only): hide the panel's welcome header under the folded caption, and in
  // Keep hide the nav column so only the Export card sits beside the caption.
  tl.set(layer, {'--mt': -60, '--ml': -60}, 0);
  tl.to(layer, {'--mt': 226, duration: 1.2, ease: 'power3.inOut'}, A(2.5));
  // MCP: drift down the capability list while it is read.
  tl.to(cam, {y: full.y - 300, duration: 2.4, ease: 'sine.inOut'}, A(11.1));
  tl.to(cam, {...full, duration: 0.6, ease: 'power2.inOut'}, A(13.75) - 0.3);
  // Keep: frame the Export card on the right half, clear of the caption and command card.
  tl.to(cam, {...frameAt(1.24, navRight + 8, navTop + 150, 940, 430), duration: 1.1, ease: 'power3.inOut'}, own.start - 0.2);
  tl.to(layer, {'--mt': 186, '--ml': 900, duration: 1.1, ease: 'power3.inOut'}, own.start - 0.2);
  // Then a slow push toward the format cards while the command is typed and Capsule is chosen.
  tl.to(cam, {...frameAt(1.3, navRight + 8, navTop + 150, 925, 445), duration: own.end - own.start - 1.4, ease: 'sine.inOut'}, own.start + 1.0);
  tl.to(layer, {'--mt': 212, duration: own.end - own.start - 1.4, ease: 'sine.inOut'}, own.start + 1.0);
  void shellLeft;

  for (const id of ['panel', 'own']) onDrive(id, local => drivePanel(local + shot(id).start));
  for (const id of ['panel', 'own']) onRender(id, (local, t) => {
    const clickTimes = CLICKS.map(c => c.at);
    paintCursor($('#panel-layer [data-cursor="panel"]'), t, t >= CLICKS[0].at - 0.9 ? cursorKeys(t) : [], clickTimes, [CLICKS[0].at - 0.9, clickTimes[clickTimes.length - 1] + 0.6]);
  });
}
