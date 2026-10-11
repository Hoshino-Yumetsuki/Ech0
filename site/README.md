# Ech0 — official site & documentation

This package is the **official** landing page and **documentation site** for [Ech0](https://github.com/lin-snow/Ech0). It is a **client-side SPA** (no SSR): React Router 7, React 19, Vite, Tailwind CSS v4, and Markdown-backed docs.

## What lives here

| Path          | Purpose                                                                                             |
| ------------- | --------------------------------------------------------------------------------------------------- |
| `app/routes/` | Pages: home (`/`), docs catalog (`/docs`), doc article (`/docs/`*), privacy                         |
| `app/docs/`   | Doc registry (`registry.ts`), Markdown rendering (`MarkdownDoc.tsx`), table of contents helpers     |
| `docs/`       | Markdown sources (`**/*.md`), images under `docs/imgs/` (copied into `public/docs-assets` on build) |
| `public/`     | Static assets (`logo.svg`, `screenshot.png`, `_redirects` for SPA fallback on static hosts)         |

The doc catalog order and featured cards are controlled in `app/docs/registry.ts` (`DOC_ORDER`, optional `DOC_HERO_SLUGS`).

## Prerequisites

- **Node.js** (LTS recommended)
- **pnpm** (workspace uses `pnpm` at the repo root)

## Commands

```bash
pnpm install
pnpm dev          # dev server (Vite; default http://localhost:5173)
pnpm build        # output: build/client/ (static assets)
pnpm start        # serve build/client locally (serve -s)
pnpm typecheck    # react-router typegen + tsc
pnpm lint
pnpm format       # or pnpm format:check
```

`prebuild` copies `docs/imgs` → `public/docs-assets` when present, so Markdown can reference `/docs-assets/imgs/...`.

## Environment

| Variable        | Purpose                                                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `VITE_SITE_URL` | Canonical site origin (no trailing slash) for OG URLs, JSON-LD, sitemap-related logic. Default in code: `https://www.ech0.app`. |

When deploying to a custom domain, align `VITE_SITE_URL` and update `public/sitemap.xml` / `public/robots.txt` if needed.

## Production build

- Run `pnpm build`.
- Deploy the contents of `**build/client/**` to any static host.
- Configure **SPA fallback** to `index.html` for client-side routes (this repo includes `public/_redirects` for Netlify-style hosts).

## Editing documentation

1. Add or change Markdown under `docs/` (e.g. `docs/guide/foo.md` → URL `/docs/guide/foo`).
2. Register the slug in `app/docs/registry.ts` `DOC_ORDER` if you care about sidebar order; unlisted files still appear but sort after known entries.
3. Use `![](imgs/...)` in Markdown; paths are rewritten to `/docs-assets/imgs/...` at render time.

## License

Content and code follow the same terms as the parent [Ech0](https://github.com/lin-snow/Ech0) repository unless noted otherwise.
