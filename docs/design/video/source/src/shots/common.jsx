// Shared film pieces derived from Ech0 itself (DIRECTION.md §4):
// - Rings: echo rings from a date dot (the product is named Echo; every timeline entry starts with a dot)
// - typeInto: typed text with the product's own typewriter cursor (HomeHeader .home-header__cursor)
import React from 'react';
import {shot} from '../engine.js';

export const sec = id => document.querySelector(`section[data-shot="${id}"]`);
export const q = (id, sel) => sec(id)?.querySelector(sel);
export const $ = sel => document.querySelector(sel);
export const BEAT = 60 / 96;
export const showShot = (tl, id) => { const s = shot(id); tl.set(sec(id), {opacity: 1}, s.start); tl.set(sec(id), {opacity: 0}, s.end); };

/** Concentric thin rings, radius driven through the --r custom property. */
export function Rings({name, count = 3, className = ''}) {
  return <div className={'rings ' + className} data-rings={name}>
    {Array.from({length: count}, (_, i) => <i key={i} className="ring" />)}
  </div>;
}
/** Pulse rings of `root` outward: ring i starts at `at + i*gap`, grows to `maxR` over `dur`. */
export function pulseRings(tl, root, at, {gap = BEAT, maxR = 900, dur = 2.2, from = 0.9, count} = {}) {
  const rings = [...root.querySelectorAll('.ring')].slice(0, count ?? 99);
  rings.forEach((ring, i) => {
    tl.fromTo(ring, {'--r': 6, opacity: from}, {'--r': maxR, opacity: 0, duration: dur, ease: 'power2.out', immediateRender: false}, at + i * gap);
  });
}
/** Ring radius at time t for one ring pulse (same curve as pulseRings), for hit-testing. */
export const ringRadius = (t, start, maxR, dur = 2.2) => {
  if (t < start) return -1;
  const u = Math.min(1, (t - start) / dur);
  return 6 + (maxR - 6) * (1 - (1 - u) * (1 - u));
};

/** Type `text` into `el` from `at` at `cps` chars/s; the cursor blinks (product cadence 0.95s) after typing. */
export function typeInto(el, local, at, text, cps, {cursorUntil = Infinity, cursorFrom = at - 0.6} = {}) {
  if (!el) return;
  const n = local < at ? 0 : Math.min(text.length, Math.floor((local - at) * cps) + 1);
  const t = el.querySelector('.typed-text');
  const next = text.slice(0, n);
  if (t.textContent !== next) t.textContent = next;
  const done = n >= text.length;
  const cursor = el.querySelector('.typed-cursor');
  const typing = local >= at && !done;
  const blinkOn = typing || Math.floor((local - cursorFrom) / 0.475) % 2 === 0;
  cursor.style.opacity = local >= cursorFrom && local < cursorUntil && blinkOn ? '1' : '0';
}
export const Typed = ({className = ''}) => <span className={'typed ' + className}><span className="typed-text" /><span className="typed-cursor" /></span>;
