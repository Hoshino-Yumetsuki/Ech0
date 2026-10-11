import React from 'react';
import {ShotContext} from './film-store.js';
import {shots} from './engine.js';
import {SHOT_VIEWS} from './shots/index.js';
import {HomeLayer} from './shots/home-layer.jsx';
import {Rail} from './shots/rail.jsx';
import {PanelLayer} from './shots/panel.jsx';
import {DetailLayer} from './shots/detail.jsx';
// Layer order: paper → real Ech0 home (write → land → stream → reach) → real admin panel (panel → own)
// → film rail (timeline-entry captions) → per-shot sections (title cards, Copilot, close).
export function Film() {
  return <main id="film" className="film-theme">
    <div className="paper" />
    <HomeLayer />
    <DetailLayer />
    <PanelLayer />
    <Rail />
    {shots.map((s, i) => {
      const View = SHOT_VIEWS[s.id];
      if (!View) throw new Error(`No view for shot "${s.id}" in src/shots/index.js`);
      return <section key={s.id} data-shot={s.id} className={'shot shot-' + s.id} style={{zIndex: 30 + i}}>
        <ShotContext.Provider value={s}><View shot={s} /></ShotContext.Provider>
      </section>;
    })}
  </main>;
}
