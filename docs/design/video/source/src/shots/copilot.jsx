// The real Ech0 Copilot chat (ChatPage → TheChatBox). The question is typed into the real
// composer and sent with the real Send button; the answer arrives as real SSE frames
// (searching → sources → coverage → delta… → done) released by film time from src/fake-api.js.
import React from 'react';
import {VueHost} from '../kit/vue-host.jsx';
import {setFieldValue} from '../kit/interact.js';
import {shot, onDrive, frame} from '../engine.js';
import {ECHOS, NEW_ECHO} from '../fixtures/api.js';
import {showShot, q} from './common.jsx';

export const QUESTION = 'What have I been writing about lately?';
// Plain prose on purpose: while streaming, AnimatedMarkdown wraps the text after an inline **/_ span
// as one inline-block that jumps to a new line (settles once done) — kept out of frame.
const ANSWER = 'Lately you keep coming back to owning your words. You moved your notes off the feed and onto your own server, and wrote yourself a short rule: write for yourself first, publish if you like. In between, you caught the harbor at dawn and the ridge at last light.';
const src = (e, distance) => ({echo_id: e.id, content: e.content, username: e.username, echo_created: e.created_at, distance,
  files: (e.echo_files || []).map(f => ({id: f.file.id, key: f.file.key, storage_type: f.file.storage_type, url: f.file.url, content_type: f.file.content_type, category: f.file.category, width: f.file.width, height: f.file.height}))});
const byId = id => [NEW_ECHO, ...ECHOS].find(e => e.id === id);

export function CopilotView() {
  return <div className="copilot-stage"><div className="chat-cam"><VueHost name="chat" className="chat-host" /></div></div>;
}

export function buildCopilot(tl) {
  const s = shot('copilot');
  showShot(tl, 'copilot');
  const stage = q('copilot', '.copilot-stage');
  tl.fromTo(stage, {opacity: 0, y: 60}, {opacity: 1, y: 0, duration: 0.9, ease: 'expo.out', immediateRender: false}, s.start - 0.05);
  tl.set(stage, {opacity: 0}, 0);
  tl.to(stage, {opacity: 0, duration: 0.01}, s.end);
  const cam = q('copilot', '.chat-cam');
  // Chat column is page x≈624–1295; frame it on the right half, clear of the rail captions.
  tl.set(cam, {transformOrigin: '0 0', scale: 1.36, x: 910 - 1.36 * 624, y: -150}, 0);
  tl.to(cam, {y: -30, duration: 1.0, ease: 'power2.inOut'}, s.start + 2.6);

  const TYPE_AT = 0.8, CPS = 24, SEND = 2.5;
  const words = ANSWER.split(/(?<=\s)/);
  const events = [
    {at: 2.75, name: 'searching', data: {name: 'search_echos', query: 'recent posts: notes, photos, links'}},
    {at: 3.25, name: 'searching', data: {name: 'search_echos', query: 'own server, self-hosting'}},
    {at: 3.75, name: 'sources', data: [src(byId('e-new'), 0.18), src(byId('e-quiet'), 0.24), src(byId('e-harbor'), 0.31), src(byId('e-ridge'), 0.37)]},
    {at: 3.85, name: 'coverage', data: {total: 6, returned: 4, buckets: 2, truncated: false}},
    ...words.map((w, i) => ({at: 4.25 + i * 0.085, name: 'delta', data: {text: w}})),
  ];
  events.push({at: events[events.length - 1].at + 0.3, name: 'done', data: {done: true}});

  let last = -1, sentUpTo = 0;
  const host = () => q('copilot', '.chat-host');
  onDrive('copilot', async local => {
    if (local === last) return;
    if (local < last && window.__chatFeed) {
      // Back-seek: remount a fresh chat so the conversation replays from nothing.
      window.__chatFeed = null; sentUpTo = 0;
      const b = window.Ech0Bridge, el = host();
      el.__handle?.unmount(); el.innerHTML = '';
      el.__handle = await b.mount(el, 'chat');
      await frame();
    }
    last = local;
    const field = host()?.querySelector('.composer__field');
    if (!window.__chatFeed) {
      const n = local < TYPE_AT ? 0 : Math.min(QUESTION.length, Math.floor((local - TYPE_AT) * CPS) + 1);
      setFieldValue(field, QUESTION.slice(0, n));
      if (local >= SEND && field?.value === QUESTION) {
        await frame();
        host().querySelector('.composer__action--send')?.click();
        for (let i = 0; i < 30 && !window.__chatFeed; i++) await new Promise(r => setTimeout(r, 10));
      }
    }
    const feed = window.__chatFeed;
    if (!feed) return;
    let pushed = false;
    while (sentUpTo < events.length && events[sentUpTo].at <= local) {
      const e = events[sentUpTo++];
      if (e.name === 'done') { feed.push('done', e.data); feed.close(); } else feed.push(e.name, e.data);
      pushed = true;
    }
    if (pushed) { await new Promise(r => setTimeout(r, 30)); await frame(); }
  });
}
