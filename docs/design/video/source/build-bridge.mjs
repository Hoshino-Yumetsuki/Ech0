// Builds the Vue "product bridge": Ech0's real Vue components, compiled with the product's own
// Vite + @vitejs/plugin-vue + UnoCSS + Sass toolchain (resolved from the product's node_modules),
// wrapped in an IIFE that exposes window.Ech0Bridge for the React film engine to mount per shot.
// Nothing in the product repo is written: output lands in .build/bridge/.
import path from 'node:path';
import {pathToFileURL} from 'node:url';
import integration from './integration.config.mjs';

const video = path.resolve('.');
const web = path.join(integration.repoDir, 'web');
const nm = p => path.join(web, 'node_modules', p);
const imp = async (p, file) => import(pathToFileURL(path.join(nm(p), file)).href);

const {build} = await imp('vite', 'dist/node/index.js');
const {default: vue} = await imp('@vitejs/plugin-vue', 'dist/index.mjs');
const {default: UnoCSS} = await imp('unocss', 'dist/vite.mjs');

await build({
  root: web,                 // UnoCSS picks up the product's uno.config.ts; envDir -> product .env*
  mode: 'production',
  configFile: false,
  publicDir: path.join(web, 'public'),   // resolves /Ech0.svg-style imports; copied into dist/ by build.mjs
  logLevel: 'warn',
  plugins: [
    {
      // Bare imports from film-side files resolve as if written inside the product (its node_modules).
      name: 'resolve-from-product', enforce: 'pre',
      async resolveId(source, importer, opts) {
        if (!importer || !importer.startsWith(video) || /^[./]|^\0|^virtual:|^@\//.test(source)) return null;
        return this.resolve(source, path.join(web, 'src/main.ts'), {...opts, skipSelf: true});
      },
    },
    vue({template: {compilerOptions: {isCustomElement: tag => tag === 'meting-js' || tag === 'cap-widget'}}}),
    UnoCSS({
      configFile: path.join(web, 'uno.config.ts'),
      content: {filesystem: [path.join(video, 'src/vue/**/*.vue')]},
    }),
  ],
  resolve: {
    alias: [
      {find: /^@\//, replacement: path.join(web, 'src') + '/'},
      // Film-side Vue files live outside the product tree: point shared runtimes at the product's copies.
    ],
    dedupe: ['vue', 'pinia', 'vue-router', 'vue-i18n'],
  },
  define: {'process.env.NODE_ENV': '"production"'},
  build: {
    outDir: path.join(video, '.build/bridge'),
    emptyOutDir: true,
    copyPublicDir: false,
    minify: true,
    cssCodeSplit: false,
    reportCompressedSize: false,
    lib: {entry: path.join(video, 'src/vue/bridge.ts'), name: 'Ech0Bridge', formats: ['iife'], fileName: () => 'bridge.js', cssFileName: 'bridge'},
    rollupOptions: {output: {assetFileNames: 'assets/[name]-[hash][extname]'}},
  },
});
console.log('Built .build/bridge/bridge.js');
