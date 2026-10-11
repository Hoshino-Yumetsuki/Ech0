import http from 'node:http';
import path from 'node:path';
import {readFile, realpath} from 'node:fs/promises';
const types = {'.mp4': 'video/mp4', '.html': 'text/html', '.css': 'text/css', '.js': 'text/javascript', '.svg': 'image/svg+xml', '.png': 'image/png', '.jpg': 'image/jpeg', '.webp': 'image/webp', '.woff': 'font/woff', '.woff2': 'font/woff2', '.ttf': 'font/ttf', '.json': 'application/json', '.wav': 'audio/wav', '.mp3': 'audio/mpeg'};
export async function serve(port = 0) {
  const base = await realpath('dist');
  const server = http.createServer(async (req, res) => {
    try {
      const name = decodeURIComponent(new URL(req.url, 'http://localhost').pathname);
      const file = await realpath(path.join(base, name === '/' ? 'index.html' : name));
      if (file !== base && !file.startsWith(base + path.sep)) { res.writeHead(403); res.end(); return; }
      const body = await readFile(file), type = types[path.extname(file)] || 'application/octet-stream';
      // Byte ranges: <video> is only seekable when the server honours Range requests.
      const m = /bytes=(\d*)-(\d*)/.exec(req.headers.range || '');
      if (m) {
        const start = m[1] ? +m[1] : body.length - +m[2], end = m[1] && m[2] ? Math.min(+m[2], body.length - 1) : body.length - 1;
        res.writeHead(206, {'Content-Type': type, 'Accept-Ranges': 'bytes', 'Content-Range': `bytes ${start}-${end}/${body.length}`, 'Content-Length': end - start + 1});
        res.end(body.subarray(start, end + 1)); return;
      }
      res.writeHead(200, {'Content-Type': type, 'Accept-Ranges': 'bytes', 'Content-Length': body.length}); res.end(body);
    } catch { res.writeHead(404); res.end('Not found'); }
  });
  await new Promise((resolve, reject) => { server.once('error', reject); server.listen(port, '127.0.0.1', resolve); });
  return {server, url: `http://127.0.0.1:${server.address().port}`};
}
// Open the film under the demo instance's origin (https://mira.example): product UI that prints
// window.location.origin (e.g. the MCP endpoint) shows the demo domain instead of 127.0.0.1:port.
export const FILM_ORIGIN = 'https://mira.example';
export async function openFilm(page, localUrl) {
  await page.route(FILM_ORIGIN + '/**', async route => {
    const u = new URL(route.request().url());
    const response = await route.fetch({url: localUrl + u.pathname + u.search});
    await route.fulfill({response});
  });
  await page.goto(FILM_ORIGIN + '/', {waitUntil: 'networkidle'});
}
