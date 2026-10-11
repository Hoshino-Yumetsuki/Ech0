// The film's own timeline: one orange dot + a rail, with each chapter caption laid out like an
// Ech0 timeline entry (dot · accent label · content). The dot is the same dot from the opening.
import React from 'react';
import {Split, reveal} from '../kit/text.jsx';
import {shot, shots} from '../engine.js';
import {Rings, pulseRings, Typed, typeInto, BEAT, $} from './common.jsx';
import {onRender} from '../engine.js';

const LABELS = {write: 'Write', land: 'Publish', detail: 'Share', stream: 'Read', reach: 'Reach', copilot: 'Ask', panel: 'Manage', own: 'Keep'};
const ENTRY_IDS = Object.keys(LABELS);
export const DOT = {cx: 574, cy: 540, rx: 150, ry: 372};   // opening centre → rail anchor

export function Rail() {
  return <div id="rail">
    <div className="rail-line" />
    <div className="film-dot"><Rings name="dot" count={3} /><b /></div>
    <div className="open-line"><Typed className="open-say" /><span className="open-echo"><Split text="Let it echo." /></span></div>
    {ENTRY_IDS.map(id => {
      const s = shot(id);
      return <div key={id} className="entry" data-entry={id}>
        <div className="entry-label">{LABELS[id]}</div>
        <h2 className="entry-title">{s.headlineEn.split('\n').map((l, i) => <span key={i} className="entry-line"><Split text={l} /></span>)}</h2>
        {s.description ? <p className="entry-sub">{s.description}</p> : null}
      </div>;
    })}
  </div>;
}

export function buildRail(tl) {
  const dot = $('#rail .film-dot'), core = $('#rail .film-dot b'), line = $('#rail .rail-line');
  const open = shot('open'), echo = shot('echo'), close = shot('close');
  // Opening: the dot lights on the first beat, then the sentence types beside it.
  tl.set(dot, {x: DOT.cx, y: DOT.cy}, 0);
  tl.fromTo(core, {scale: 0}, {scale: 1, duration: 0.55, ease: 'expo.out'}, open.start + BEAT);
  const say = $('#rail .open-say');
  onRender('open', local => typeInto(say, local, 2 * BEAT, 'Say it once.', 11, {cursorFrom: BEAT + 0.3, cursorUntil: 5.0 - open.start}));
  onRender('echo', local => typeInto(say, local + (echo.start - open.start), 2 * BEAT, 'Say it once.', 11, {cursorFrom: BEAT + 0.3, cursorUntil: 5.0 - open.start}));
  pulseRings(tl, $('#rail [data-rings="dot"]'), open.start + 5 * BEAT, {count: 1, maxR: 520, dur: 1.9});
  // "Let it echo.": the first sentence blurs away, three rings spread past the frame.
  tl.to(say, {opacity: 0, filter: 'blur(10px)', y: -20, duration: 0.45, ease: 'power2.in'}, echo.start - 0.1);
  reveal(tl, $('#rail .open-echo'), echo.start + 0.3, {stagger: 0.03});
  pulseRings(tl, $('#rail [data-rings="dot"]'), echo.start, {count: 3, maxR: 1500, dur: 2.4});
  tl.to($('#rail .open-echo'), {opacity: 0, filter: 'blur(10px)', y: -20, duration: 0.45, ease: 'power2.in'}, echo.start + 1.85);
  // The dot slides to the rail anchor and grows the rail: the film's own timeline.
  tl.to(dot, {x: DOT.rx, y: DOT.ry, duration: 1.1, ease: 'power3.inOut'}, echo.start + 2.0);
  tl.fromTo(line, {scaleY: 0}, {scaleY: 1, duration: 1.0, ease: 'power2.inOut'}, echo.start + 2.6);

  // Chapter entries arrive like new Echos: from below, with a ring from the dot.
  ENTRY_IDS.forEach((id, k) => {
    const s = shot(id), el = $(`#rail [data-entry="${id}"]`);
    const inAt = s.start + (k === 0 ? 0.1 : 0.0);
    tl.fromTo(el, {y: 150, opacity: 0}, {y: 0, opacity: 1, duration: 0.85, ease: 'expo.out', immediateRender: false}, inAt);
    reveal(tl, el.querySelector('.entry-title'), inAt + 0.1, {stagger: 0.022, dur: 0.6});
    const sub = el.querySelector('.entry-sub');
    if (sub) tl.fromTo(sub, {y: 18, opacity: 0}, {y: 0, opacity: 1, duration: 0.6, ease: 'expo.out', immediateRender: false}, inAt + s.descriptionAt);
    if (k > 0) pulseRings(tl, $('#rail [data-rings="dot"]'), inAt, {count: 1, maxR: 360, dur: 1.4});
    tl.set(el, {opacity: 0}, 0);
    if (id === 'panel') {
      // Manage: after the caption is read, it folds into a page heading so the panel can fill the frame.
      const fold = s.start + 2.5;
      tl.to(el, {y: -296, scale: 0.5, transformOrigin: '0% 0%', duration: 1.2, ease: 'power3.inOut'}, fold);
      if (sub) tl.to(sub, {opacity: 0, duration: 0.4, ease: 'power2.in'}, fold);
      tl.to($('#rail .film-dot'), {y: 66, duration: 1.2, ease: 'power3.inOut'}, fold);
      tl.to(el, {y: -420, opacity: 0, duration: 0.42, ease: 'power2.in'}, s.end - 0.45);
      tl.to($('#rail .film-dot'), {y: DOT.ry, duration: 0.8, ease: 'power3.inOut'}, s.end - 0.45);
      return;
    }
    tl.to(el, {y: -170, opacity: 0, duration: 0.42, ease: 'power2.in'}, s.end - 0.45);
  });
  // Close: the rail folds away and the dot returns to centre to become the logo's eye.
  tl.to(line, {scaleY: 0, duration: 0.6, ease: 'power2.in'}, close.start);
}
