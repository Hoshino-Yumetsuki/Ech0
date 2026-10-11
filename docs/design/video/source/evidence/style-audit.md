# Style audit (repo style)

1. **Palette** (web/src/themes/tokens/semantic.light.scss): canvas #f4f1ec, surface #fff, text #33251d / #5b4f46 / #8e847d, border #e4dfdb / #c2b9b2, accent #b84200, accent-soft #f4e3cd. Dark (semantic.dark.scss): canvas #333. Logo dot #F54A00 (web/public/Ech0.svg).
2. **Type**: display serif `Iowan Old Style` (--font-family-display, foundation.scss) → film headlines; UI sans = system stack (--font-family-sans) → film subtitles; mono stack → command card and meta line. English-only film by user choice (typography.exceptionReason in plan.json). Fonts are macOS system fonts; verified rendered in stills (headless Chromium on macOS).
3. **Space & shape**: radii 0.25–0.875rem, soft shadows; the timeline is a hairline rail with 8px accent date dots (TheEchoCard.vue `.timeline-marker`).
4. **Brand**: Ech0.svg (rounded plate #FFF4E4 + four strokes). Close shot redraws the same paths as separate SVG groups for the assembly animation.
5. **Framing**: product UI at 1.85× (editor/timeline), 1.3→1.85× (Status), 1.36× (Copilot), 0.86→1.5× (admin panel), 1.24× (Export page). Light theme throughout (user request).
6. **Motif candidates used**: timeline date dot, echo (product name), logo strokes, HomeHeader typewriter cursor.
7. **Decision**: repo. Film captions are laid out as timeline entries (dot · accent label · content) to keep the product's own grammar.
