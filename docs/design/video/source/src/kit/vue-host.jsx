// Host element for a real Ech0 Vue composition. React owns the div; the bridge mounts a Vue app
// inside it (shared Pinia/i18n/router). client.jsx waits on window.__vueMounts before builders run.
import React, {useEffect, useRef} from 'react';
const mounts = (window.__vueMounts ??= []);
export const whenBridge = () => new Promise(r => { const c = () => (window.Ech0Bridge ? r(window.Ech0Bridge) : setTimeout(c, 5)); c(); });
export function VueHost({name, props, className, style}) {
  const ref = useRef(null);
  useEffect(() => {
    let handle;
    const p = whenBridge().then(b => b.mount(ref.current, name, props)).then(h => (handle = h));
    mounts.push(p);
    return () => { p.then(() => handle?.unmount()); };
  }, []);
  return <div ref={ref} className={className} style={style} />;
}
