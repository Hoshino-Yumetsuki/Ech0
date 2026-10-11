// Product integration for the video project (never edit the product repo).
import {fileURLToPath} from 'node:url';
// Ech0's UI is Vue 3: build-bridge.mjs compiles real product components with the product's own
// Vite toolchain into .build/bridge/, and the React film engine mounts them via window.Ech0Bridge.
export default {
  // This project lives at docs/design/video/source inside the Ech0 repo.
  repoDir: fileURLToPath(new URL('../../../../', import.meta.url)).replace(/\/$/, ''),
  aliases: {},
  external: [],
  loaders: {},
  define: {},
  esbuild: {},
};
