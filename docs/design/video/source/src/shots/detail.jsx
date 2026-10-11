// Echo detail (EchoView → EchoPage → TheEchoDetail + TheEchoInteractions/TheComment), reached by a
// real click on the new Echo's time label in the timeline. Mira shares it (real share popover →
// Copy link); then the page is shown as a signed-out reader who opened that link: they like it and
// post a reply through the real comment form. A real click on "back" returns to the timeline.
import React from 'react';
import {Cursor, paintCursor} from '../kit/cursor.jsx';
import {setFieldValue} from '../kit/interact.js';
import {shot, onDrive, onRender, frame} from '../engine.js';
import {READER, resetDetail} from '../fixtures/api.js';
import {$} from './common.jsx';

export function DetailLayer() {
  // The cursor sits outside the layer: its first click lands on the timeline, before the layer shows.
  return <>
    <div id="detail-layer">
      <div className="detail-cam"><div className="detail-host" /></div>
      <div className="detail-chip"><i />A reader opens the shared link</div>
    </div>
    <div id="detail-cursor"><Cursor id="detail" color="#33251d" stroke="#fffaf2" /></div>
  </>;
}

const B = () => window.Ech0Bridge;
const q = sel => $('#detail-layer ' + sel);
const qa = sel => [...document.querySelectorAll('#detail-layer ' + sel)];
const center = el => { const r = el.getBoundingClientRect(); return {x: r.left + r.width / 2, y: r.top + r.height / 2}; };
const until = async (cond, ms = 1500) => { const t0 = performance.now(); while (!cond() && performance.now() - t0 < ms) await new Promise(r => setTimeout(r, 16)); await frame(); };
const timeLabel = () => $('#home-layer .timeline-marker')?.nextElementSibling;
const likeBtn = () => q('.echo-meta-like');
const shareBtn = () => q('.echo-meta-actions button[aria-label]:not(.echo-meta-like)') || qa('.echo-meta-actions button').find(b => b !== likeBtn());
const copyLink = () => qa('.share-tile, button').find(b => /copy link/i.test(b.getAttribute('aria-label') || b.textContent || ''));
const pill = () => q('.comment-pill-btn');
const idInputs = () => qa('.comment-id-grid input');
const textarea = () => q('.comment-textarea');
const submitBtn = () => q('.comment-form-panel button[type="submit"]') || qa('.comment-form-panel button').pop();
const backBtn = () => q('.echo-detail-back');

export const T = {};   // absolute action times (also read by home-layer.jsx to stand aside)
let handle = null, mountedAs = null, lastT = -1, busy = null;

async function mountDetail(asVisitor) {
  const b = B(), host = q('.detail-host');
  handle?.unmount(); host.innerHTML = '';
  await b.setVisitor(asVisitor);
  handle = await b.mount(host, 'echo');
  mountedAs = asVisitor ? 'reader' : 'owner';
  await until(() => q('.echo-meta-like') && q('.comment-pill-btn'), 2000);
  await new Promise(r => setTimeout(r, 150)); await frame();
}

async function driveDetail(t) {
  if (t === lastT) return busy;
  const back = t < lastT;
  lastT = t;
  busy = (async () => {
    const b = B();
    const inDetail = t >= T.open && t < T.back;
    if (!inDetail) {
      if (t >= T.back && b.route().startsWith('/echo/')) {
        const btn = backBtn();
        if (btn && !back) btn.click(); else await b.navigate('/');
        await until(() => b.route() === '/', 1500);
      }
      if (mountedAs === 'reader' && (t < T.open || t >= T.back + 0.5)) { await b.setVisitor(false); mountedAs = 'owner-restored'; }   // after the layer has faded
      return;
    }
    if (!b.route().startsWith('/echo/')) {
      const label = timeLabel();
      if (label && !back && t - T.open < 0.5) label.click(); else await b.navigate('/echo/e-new');
      await until(() => b.route().startsWith('/echo/'), 1500);
    }
    const want = t >= T.reader ? 'reader' : 'owner';
    if (mountedAs !== want || back) { if (back) resetDetail(); await mountDetail(want === 'reader'); }
    if (want === 'owner') {
      // Share popover: open at T.share, "Copy link" at T.copy (which closes it).
      const panelOpen = () => qa('.share-tile').length > 0;
      const wantOpen = t >= T.share && t < T.copy;
      if (panelOpen() !== wantOpen) {
        if (wantOpen) shareBtn()?.click(); else copyLink()?.click() ?? shareBtn()?.click();
        await until(() => panelOpen() === wantOpen, 500);
      }
      return;
    }
    // Reader: like, open the form, type, submit.
    const liked = () => (likeBtn()?.textContent.trim() || '0') !== '0';
    if (t >= T.like && !liked()) { likeBtn()?.click(); await new Promise(r => setTimeout(r, 120)); await frame(); }
    const formOpen = () => !!textarea();
    if (t >= T.form && t < T.submit + 0.05 && !formOpen()) { pill()?.click(); await until(formOpen, 500); }
    if (formOpen() && t < T.submit) {
      const typed = (text, at, cps) => text.slice(0, Math.max(0, Math.min(text.length, Math.floor((t - at) * cps) + (t >= at ? 1 : 0))));
      const [nick, mail] = idInputs();
      setFieldValue(nick, typed(READER.nickname, T.nick, 14));
      setFieldValue(mail, typed(READER.email, T.mail, 34));
      setFieldValue(textarea(), typed(READER.content, T.text, 34));
      await frame();
    }
    if (t >= T.submit) {
      // The live preview repeats the text, so count comments instead of matching it.
      const posted = () => /\b2 comments\b/.test(q('.echo-interactions')?.textContent || '');
      if (!posted() && textarea()?.value === READER.content) { submitBtn()?.click(); await until(posted, 1500); await new Promise(r => setTimeout(r, 120)); }
    }
  })();
  return busy;
}

export function buildDetail(tl) {
  const d = shot('detail');
  const A = rel => d.start + rel;
  Object.assign(T, {open: A(0.625), share: A(1.875), copy: A(3.125), reader: A(3.75), like: A(4.375), form: A(5.0),
    nick: A(5.5), mail: A(5.95), text: A(6.6), submit: A(8.75), back: A(11.875)});
  const layer = $('#detail-layer'), cam = $('#detail-layer .detail-cam'), chip = $('#detail-layer .detail-chip');
  tl.set(layer, {opacity: 0}, 0);
  tl.fromTo(layer, {opacity: 0}, {opacity: 1, duration: 0.35, ease: 'power1.out', immediateRender: false}, T.open + 0.05);
  // Reader view: a brief blink as the page reloads signed-out, plus a chip that says what happened.
  tl.to(layer, {opacity: 0.25, duration: 0.12, ease: 'power1.in'}, T.reader - 0.12);
  tl.to(layer, {opacity: 1, duration: 0.3, ease: 'power1.out'}, T.reader);
  tl.fromTo(chip, {opacity: 0, y: -10}, {opacity: 1, y: 0, duration: 0.5, ease: 'expo.out', immediateRender: false}, T.reader + 0.05);
  tl.set(chip, {opacity: 0}, 0);
  tl.to(chip, {opacity: 0, duration: 0.3}, T.back - 0.3);
  tl.to(layer, {opacity: 0, duration: 0.35, ease: 'power1.in'}, T.back + 0.05);
  // Camera: the detail column (page x≈768–1152) at 2× on the right half. The page is taller than the
  // frame (the Echo carries a video), so the camera follows whatever is being clicked: each focus is a
  // live element measured in page space (offset chain, unaffected by the camera's own transform).
  const S2 = 2.0, QX = 1330, PX = 960;
  const pageY = el => { let y = 0, n = el; const host = q('.detail-host'); while (n && n !== host) { y += n.offsetTop; n = n.offsetParent; } return y; };
  const newReply = () => [...document.querySelectorAll('#detail-layer .echo-interactions *')].find(el => el.children.length && el.offsetHeight < 260 && /#2/.test(el.textContent) && el.textContent.includes('Jonas') && !el.querySelector('textarea'));
  const FOCUS = [
    {at: T.open, el: () => q('.echo-detail-head'), qy: 90},
    {at: T.share - 0.75, el: () => q('.echo-meta-line'), qy: 560},
    {at: T.form + 0.85, el: pill, qy: 230},             // moves only after the pill click (T.form)
    {at: T.submit + 0.9, el: () => newReply() || q('.comment-form-panel'), qy: 600},   // after the submit click; held ≈1 s
    {at: T.back - 0.45, el: () => q('.echo-detail-head'), qy: 90},     // up to the header just before back
  ];
  const camY = focus => { const el = focus.el(); return el ? focus.qy - S2 * pageY(el) : null; };
  const ease = u => (u < 0.5 ? 4 * u * u * u : 1 - Math.pow(-2 * u + 2, 3) / 2);
  const placeCamera = t => {
    let i = 0;
    while (i + 1 < FOCUS.length && t >= FOCUS[i + 1].at) i++;
    let y = camY(FOCUS[i]) ?? 110 - S2 * 40;
    const next = FOCUS[i + 1], MOVE = 0.8;
    if (next && t > next.at - MOVE) { const y2 = camY(next); if (y2 !== null) y += (y2 - y) * ease((t - (next.at - MOVE)) / MOVE); }
    cam.style.transformOrigin = '0 0';
    cam.style.transform = `translate(${QX - S2 * PX}px, ${y}px) scale(${S2})`;
  };

  for (const id of ['land', 'detail', 'stream']) onDrive(id, local => driveDetail(local + shot(id).start));
  const clicks = [
    {at: T.open, el: timeLabel}, {at: T.share, el: shareBtn}, {at: T.copy, el: () => copyLink() || shareBtn()},
    {at: T.like, el: likeBtn}, {at: T.form, el: pill}, {at: T.submit, el: submitBtn}, {at: T.back, el: backBtn},
  ];
  onRender('detail', (local, t) => {
    placeCamera(t);   // camera first, then the pointer is measured against the moved page
    let prev = {x: 1760, y: 980};
    const keys = [{at: T.open - 0.9, x: prev.x, y: prev.y}];
    for (const c of clicks) { const el = c.el(); if (el) prev = center(el); keys.push({at: c.at - 0.12, x: prev.x, y: prev.y}); }
    // While the reader types, the pointer rests beside the form.
    paintCursor($('#detail-cursor [data-cursor="detail"]'), t, keys, clicks.map(c => c.at), [T.open - 0.9, T.back + 0.4]);
  });
}
