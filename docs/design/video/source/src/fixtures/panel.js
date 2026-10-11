// Admin panel fixtures (web/src/views/panel/*). Everything is fictional demo data except
// mcp-manifest.json, which is the real manifest dumped from internal/mcp (Registry.Manifest()
// with Adapter.RegisterAll) via a throwaway `go test -overlay` — see docs/design/video/README.md.
import MCP_MANIFEST from './mcp-manifest.json';

const ok = data => ({code: 1, msg: '', data});
const NOW = Math.floor(new Date('2026-09-18T09:30:00+08:00').getTime() / 1000);
const H = 3600, D = 86400;
const day = n => new Date((NOW - n * D) * 1000).toISOString().slice(0, 10);

const visitor = [[6, 412, 188], [5, 506, 231], [4, 468, 204], [3, 623, 287], [2, 571, 259], [1, 698, 315], [0, 344, 162]]
  .map(([n, pv, uv]) => ({date: day(n), pv, uv}));

const cm = (id, echo, nickname, content, hoursAgo, status, hot = false) =>
  ({id, echo_id: echo, nickname, email: `${nickname.toLowerCase()}@example.com`, website: '', content, status, hot, source: 'guest', created_at: NOW - hoursAgo * H, updated_at: NOW - hoursAgo * H});
const PANEL_COMMENTS = [
  cm('pc1', 'e-new', 'Lena', 'Same words, my address. Stealing this line for my own about page.', 0.3, 'pending'),
  cm('pc2', 'e-harbor', 'Jonas', 'Four minutes of gold is worth the alarm. Great shot.', 1, 'approved', true),
  cm('pc3', 'e-quiet', 'Aiko', 'Printing "own it either way" and taping it to my monitor.', 5, 'approved'),
  cm('pc4', 'e-link', 'Sam', 'Found you through Ech0 Hub. Subscribed via RSS!', 20, 'approved'),
  cm('pc5', 'e-ridge', 'Noor', 'Which lens was this? The layers are unreal.', 30, 'pending'),
  cm('pc6', 'e-build', 'Theo', 'Reading mode looks lovely. Any plans to open source it?', 70, 'approved'),
];

const node = (name, path, type, extra = {}) => ({name, path, node_type: type, has_children: type === 'folder', ...extra});
const f = (name, path, size, ct, daysAgo) => node(name, path, 'file', {size, content_type: ct, modified_at: NOW - daysAgo * D});
const TREE = {
  '': [node('images', 'images', 'folder'), node('audio', 'audio', 'folder'), node('snapshots', 'snapshots', 'folder'),
    f('avatar.png', 'avatar.png', 48213, 'image/png', 40)],
  images: [f('harbor-dawn.jpg', 'images/harbor-dawn.jpg', 375664, 'image/jpeg', 0), f('ridge-last-light.jpg', 'images/ridge-last-light.jpg', 263738, 'image/jpeg', 3),
    f('reading-mode.png', 'images/reading-mode.png', 182004, 'image/png', 8), f('desk-night.jpg', 'images/desk-night.jpg', 230781, 'image/jpeg', 12),
    f('market-street.jpg', 'images/market-street.jpg', 402118, 'image/jpeg', 19), f('first-snow.jpg', 'images/first-snow.jpg', 318870, 'image/jpeg', 33)],
};

const LOG_MODULES = ['echo', 'comment', 'webhook', 'embedding', 'snapshot', 'visitor', 'mcp'];
const iso = s => new Date((NOW + s) * 1000).toISOString();
export const logEntry = (offsetSec, level, module, msg, fields = {}) => ({time: iso(offsetSec), level, module, msg, fields: {module, ...fields}});
const TAIL = [
  logEntry(-612, 'info', 'server', 'http server listening', {addr: ':6277'}),
  logEntry(-610, 'info', 'task', 'scheduled snapshot task registered', {cron: '0 3 * * *'}),
  logEntry(-420, 'info', 'visitor', 'visitor stats flushed', {pv: 344, uv: 162}),
  logEntry(-300, 'info', 'embedding', 'index refreshed', {echos: 486}),
  logEntry(-122, 'info', 'comment', 'comment created', {status: 'pending', echo: 'e-new'}),
  logEntry(-120, 'info', 'webhook', 'delivered', {event: 'comment.created', status: 200, ms: 41}),
];
// Live lines pushed by the logs shot while it is on screen (src/shots/panel.jsx).
export const LIVE_LOGS = [
  [0.4, 'info', 'echo', 'echo created', {echo: 'e-new', tags: 'notes'}],
  [0.9, 'info', 'webhook', 'delivered', {event: 'echo.created', status: 200, ms: 38}],
  [1.4, 'info', 'embedding', 'embedded echo', {echo: 'e-new', dims: 1024}],
  [1.9, 'info', 'mcp', 'tools/call', {tool: 'search_posts', client: 'claude', ms: 12}],
  [2.4, 'warn', 'visitor', 'rate limit hit, request dropped', {ip: '203.0.113.7'}],
  [2.9, 'info', 'snapshot', 'export queued', {job: 'exp-0918', format: 'zip'}],
];
void LOG_MODULES;

export const PANEL_FIXTURES = {
  '/api/system/visitor-stats': ok(visitor),
  '/api/echo/page': ok({items: [], total: 486}),
  '/api/connects/health': ok(['jonas', 'aiko', 'sam', 'lena'].map((id, i) => ({id, connect_url: `https://${id}.example`, status: i === 3 ? 'offline' : 'online', version: '5.7.0'}))),
  '/api/panel/comments/settings': ok({enable_comment: true, require_approval: true, captcha_enabled: true,
    email_notify: {enabled: true, smtp_host: 'smtp.example.com', smtp_port: 465, smtp_username: 'mira', smtp_password_set: true, smtp_sender: 'mira@example.com'}}),
  '/api/panel/comments': ok({items: PANEL_COMMENTS, total: 128}),
  '/api/file/tree': url => ok({items: TREE[url.searchParams.get('prefix') || ''] || []}),
  '/api/files': ok({items: [], total: 0}),
  '/api/mcp/manifest': ok(MCP_MANIFEST),
  '/api/system/logs': ok(TAIL),
  '/api/migration/status': ok(null),
  '/api/system/check-update': ok({has_update: false, latest_version: '5.7.0'}),
};
