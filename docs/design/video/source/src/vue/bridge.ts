// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Product bridge: boots Ech0's real Vue runtime (Pinia stores, vue-i18n, the product router,
// FloatingVue, global BaseDialog, full theme + UnoCSS stylesheet) once, then mounts real product
// components into host elements that the React film engine owns. Film time is pushed in through
// tick(t) so Vue-side shot compositions stay a pure function of the timeline.
import 'virtual:uno.css'
import '@/themes/index.scss'
import 'floating-vue/dist/style.css'

import { createApp, defineComponent, h, nextTick, provide, reactive, type Component } from 'vue'
import { createPinia } from 'pinia'
import FloatingVue from 'floating-vue'
import router from '@/router'
import { viewDepthKey } from 'vue-router'
import { initStores } from '@/stores/store-init'
import { i18n, setupI18n } from '@/locales'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useEchoStore } from '@/stores/echo'
import { useEditorStore } from '@/stores/editor'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user'
import { SHOTS } from './shots'

export const film = reactive({ t: 0 })
const pinia = createPinia()
let booted: Promise<void> | null = null

export function boot(theme: 'light' | 'dark' | 'sunny' = 'light') {
  booted ??= (async () => {
    localStorage.setItem('themeMode', JSON.stringify(theme))
    localStorage.setItem('theme', JSON.stringify(theme))
    localStorage.setItem('locale', JSON.stringify('en-US'))
    // A throwaway app gives Pinia an injection context before any store is touched.
    createApp({ render: () => null }).use(pinia)
    await initStores()
    await setupI18n('en-US')
  })()
  return booted
}

type Handle = { unmount: () => void }

export async function mount(el: HTMLElement, name: string, props: Record<string, unknown> = {}): Promise<Handle> {
  await boot()
  const comp = SHOTS[name] as Component | undefined
  if (!comp) throw new Error(`Unknown bridge shot: ${name}`)
  // Route-level views (PanelView) contain their own <RouterView>: start it at depth 1 so it renders
  // the child page instead of re-rendering the matched parent (which nests PanelView in itself).
  const root = name === 'panel' ? defineComponent({ setup() { provide(viewDepthKey, 1); return () => h(comp) } }) : comp
  const app = createApp(root, props)
  app.use(pinia).use(router).use(i18n).use(FloatingVue, {
    themes: { tooltip: { triggers: [], container: 'body' } },
  })
  app.component('BaseDialog', BaseDialog)
  app.config.warnHandler = () => {}
  app.mount(el)
  await router.isReady()
  return { unmount: () => app.unmount() }
}

export async function navigate(to: string) {
  await router.push(to).catch(() => undefined)
  await nextTick()
}

export async function tick(t: number) {
  if (film.t !== t) film.t = t
  await nextTick()
}

/** Real theme switch through the product's ThemeStore (html.light / html.dark / html.sunny). */
export async function setTheme(mode: 'light' | 'dark' | 'sunny') {
  const theme = useThemeStore()
  if (theme.mode === mode && document.documentElement.classList.contains(mode)) return
  theme.mode = mode
  theme.applyTheme()
  await nextTick()
}

export const route = () => router.currentRoute.value.fullPath

/** Detail chapter: view the page as a signed-out reader who opened a shared link (and back). */
let owner: unknown = null
export async function setVisitor(on: boolean) {
  const users = useUserStore()
  if (on && users.user) { owner = users.user; users.user = null }
  if (!on && !users.user && owner) users.user = owner as typeof users.user
  await nextTick()
}

/** Back-seek support: undo a publish so the write shot can replay from a clean editor. */
export async function resetEditor() {
  useEditorStore().clearEditor()
  await nextTick()
}

export async function refreshEchos() {
  await useEchoStore().refreshEchos()
  await nextTick()
}

export { i18n }
