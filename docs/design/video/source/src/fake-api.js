// Deterministic browser runtime for the film. Import before any product module.
// - fixed clock and seeded Math.random (time-of-day greetings, random picks)
// - fixture-backed fetch for the product's API routes, so real components run their effects offline
// - silent EventSource / WebSocket stubs (no streaming connections)
import {API_FIXTURES, API_PREFIXES} from './fixtures/api.js';

const FIXED_NOW = new Date('2026-09-18T09:30:00+08:00').getTime();   // a weekday morning; matches fixtures/api.js
const RealDate = Date;
class FilmDate extends RealDate { constructor(...a) { a.length ? super(...a) : super(FIXED_NOW); } static now() { return FIXED_NOW; } }
globalThis.Date = FilmDate;
let seed = 7418;
Math.random = () => { seed = (seed * 16807) % 2147483647; return (seed - 1) / 2147483646; };
try { localStorage.clear(); sessionStorage.clear(); } catch {}

window.__filmApiLog = [];
const json = (data, status = 200) => new Response(JSON.stringify(data), {status, headers: {'Content-Type': 'application/json'}});
const realFetch = window.fetch.bind(window);
window.fetch = async (input, init) => {
  const url = new URL(typeof input === 'string' ? input : input.url, location.href);
  if (url.origin !== location.origin) { window.__filmApiLog.push('EXTERNAL ' + url.href); return json({}, 404); }   // offline film
  if (!API_PREFIXES.some(p => url.pathname.startsWith(p))) return realFetch(input, init);
  window.__filmApiLog.push(url.pathname + url.search);   // inspect in stills.mjs output to find missing fixtures
  if (url.pathname === '/api/chat' && (init?.method || 'GET').toUpperCase() === 'POST') return chatStreamResponse();
  const h = API_FIXTURES[url.pathname];   // (query strings are matched by pathname only)
  return h === undefined ? json({}) : json(typeof h === 'function' ? h(url, init) : h);
};
// Copilot SSE (/api/chat): an open stream whose events the copilot shot's driver releases by film
// time (src/shots/copilot.jsx), so the real TheChatBox parses real SSE frames deterministically.
function chatStreamResponse() {
  const enc = new TextEncoder();
  let ctrl;
  const body = new ReadableStream({start: c => { ctrl = c; }});
  window.__chatFeed = {sent: 0, closed: false,
    push(name, data) { ctrl.enqueue(enc.encode(`event: ${name}\ndata: ${JSON.stringify(data)}\n\n`)); },
    close() { if (!this.closed) { this.closed = true; ctrl.close(); } }};
  return new Response(body, {status: 200, headers: {'Content-Type': 'text/event-stream'}});
}
// Ech0 honours prefers-reduced-motion (AnimatedMarkdown, ChatActivity, theme view transitions):
// report it so wall-clock reveal loops step aside and the film clock paces everything.
const realMatchMedia = window.matchMedia.bind(window);
window.matchMedia = q => (/prefers-reduced-motion:\s*reduce/.test(q)
  ? {matches: true, media: q, onchange: null, addEventListener() {}, removeEventListener() {}, addListener() {}, removeListener() {}, dispatchEvent: () => false}
  : realMatchMedia(q));
class Silent { constructor() { this.readyState = 0; } close() {} send() {} addEventListener() {} removeEventListener() {} }
window.EventSource = Silent;
// System log live stream (/ws/system/logs): opens immediately; the logs shot pushes entries by film
// time through window.__logFeed, framed exactly like the backend ({code, msg, data: Entry}).
class FilmSocket extends Silent {
  constructor(url) {
    super();
    this.url = String(url);
    if (!this.url.includes('/ws/system/logs')) return;
    window.__logFeed = this;
    setTimeout(() => { this.readyState = 1; this.onopen?.({type: 'open'}); }, 0);
  }
  push(entry) { if (this.readyState === 1) this.onmessage?.({data: JSON.stringify({code: 1, msg: '', data: entry})}); }
  close() { this.readyState = 3; this.onclose?.({type: 'close'}); }
}
window.WebSocket = FilmSocket;
// Local file upload (web/src/lib/file/upload.ts: XHR POST /api/files/upload with FormData). The
// write shot's driver advances progress by film time through window.__uploadFeed and then
// answers with the backend's success envelope.
const RealXHR = window.XMLHttpRequest;
class FilmXHR {
  constructor() { this.upload = {}; this.readyState = 0; this.status = 0; this.responseText = ''; this.headers = {}; }
  open(method, url) { this.method = method; this.url = String(url); this.readyState = 1; }
  setRequestHeader(k, v) { this.headers[k] = v; }
  abort() { this.onabort?.(); }
  send(body) {
    if (!this.url.includes('/api/files/upload')) { const x = new RealXHR(); x.open(this.method, this.url); x.onload = () => { this.status = x.status; this.responseText = x.responseText; this.onload?.(); }; x.send(body); return; }
    window.__filmApiLog?.push('XHR ' + this.url);
    const file = body instanceof FormData ? body.get('file') : body;
    const total = file?.size || 1;
    window.__uploadFeed = {
      total, done: false,
      progress: p => this.upload.onprogress?.({lengthComputable: true, loaded: Math.round(total * p), total}),
      finish: data => { if (window.__uploadFeed.done) return; window.__uploadFeed.done = true; this.readyState = 4; this.status = 200; this.responseText = JSON.stringify({code: 1, msg: '', data}); this.onload?.(); },
    };
  }
}
window.XMLHttpRequest = FilmXHR;
