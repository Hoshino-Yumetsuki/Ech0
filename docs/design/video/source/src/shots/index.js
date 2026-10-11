// Map every plan.json shot id to a view and a builder. Builders receive the master timeline.
// rail/home/panel are persistent layers (src/film.jsx) with their own builders.
import {buildRail} from './rail.jsx';
import {buildHome} from './home-layer.jsx';
import {buildPanel} from './panel.jsx';
import {buildDetail} from './detail.jsx';
import {CopilotView, buildCopilot} from './copilot.jsx';
import {Empty, OwnView, buildOwn, CloseView, buildClose} from './titles.jsx';
import {showShot} from './common.jsx';
const show = id => tl => showShot(tl, id);
export const SHOT_VIEWS = {open: Empty, echo: Empty, write: Empty, land: Empty, detail: Empty, stream: Empty, reach: Empty, copilot: CopilotView, panel: Empty, own: OwnView, close: CloseView};
export const SHOT_BUILDERS = {
  rail: buildRail, detailLayer: buildDetail, home: buildHome, panelLayer: buildPanel,
  open: show('open'), echo: show('echo'), write: show('write'), land: show('land'), detail: show('detail'), stream: show('stream'), reach: show('reach'), panel: show('panel'),
  copilot: buildCopilot, own: buildOwn, close: buildClose,
};
