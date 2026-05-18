<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="home-page">
    <div class="home-shell">
      <!-- 左栏：品牌 + 主导航 + 快捷操作（≥820px 显示） -->
      <HomeLeftRail class="home-shell__left" />

      <!-- 中栏：主信息流 / 编辑器 / 标签 / 广场 / 状态 -->
      <main
        ref="mainColumn"
        class="home-shell__main"
        :class="{ 'home-shell__main--unclipped': activeTab === 'tags' }"
      >
        <!-- 移动端品牌 + 导航：仅 <820px 显示 -->
        <div class="home-shell__mobile-top">
          <HomeHeader class="home-shell__mobile-header" />
          <HomeSidebarNav
            v-model:mobile-search-open="mobileSearchOpen"
            class="home-shell__mobile-nav"
            @open-palette="paletteOpen = true"
          />
        </div>

        <div
          v-if="activeTab === 'publish'"
          class="home-content-block home-content-block--publish"
        >
          <TheEditor />
        </div>
        <div v-else-if="activeTab === 'tags'" class="home-content-block">
          <TheTagsManager />
        </div>

        <template v-else-if="activeTab === 'home'">
          <HomeBanner :class="{ 'home-banner--mobile-hidden': shouldHideBannerOnMobile }" />
          <TheEchos :scroll-target="mainColumn" />
        </template>

        <div v-else-if="activeTab === 'status'" class="home-content-block home-status-widgets">
          <PanelCard border-style="solid" class="home-status-widgets__card">
            <TheHeatMap />
          </PanelCard>
          <PanelCard
            v-if="AgentSetting.enable"
            border-style="solid"
            class="home-status-widgets__card"
          >
            <TheRecentCard />
          </PanelCard>
          <PanelCard border-style="solid" class="home-status-widgets__card">
            <TheConnectWidget />
          </PanelCard>
          <PanelCard border-style="solid" class="home-status-widgets__card">
            <TheCommentWidget />
          </PanelCard>
        </div>

        <HubPage v-else embedded :scroll-target="mainColumn" />
      </main>

      <!-- 右栏：搜索 + 常驻 widgets + version（≥1100px 显示） -->
      <HomeRightRail class="home-shell__right" @open-palette="paletteOpen = true" />
    </div>
    <TheCommandPalette v-model="paletteOpen" />
  </div>
</template>

<script setup lang="ts">
import HomeHeader from './HomeHeader.vue'
import HomeBanner from './HomeBanner.vue'
import HomeSidebarNav from './HomeSidebarNav.vue'
import HomeLeftRail from './HomeLeftRail.vue'
import HomeRightRail from './HomeRightRail.vue'
import TheEchos from './TheEchos.vue'
import TheCommandPalette from './TheCommandPalette.vue'
import PanelCard from '@/layout/PanelCard.vue'
import { defineAsyncComponent, onMounted, ref, onBeforeUnmount, computed, watch } from 'vue'
import { useEchoStore, useUserStore, useSettingStore } from '@/stores'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import {
  TheCommentWidget,
  TheConnectWidget,
  TheHeatMap,
  TheRecentCard,
} from '@/components/advanced/widget'

const route = useRoute()
const TheEditor = defineAsyncComponent(() => import('./TheEditor.vue'))
const TheTagsManager = defineAsyncComponent(() => import('./TheEditor/TheTagsManager.vue'))
const HubPage = defineAsyncComponent(() => import('@/views/hub/modules/HubPage.vue'))

const userStore = useUserStore()
const settingStore = useSettingStore()
const echoStore = useEchoStore()
const { isLogin } = storeToRefs(userStore)
const { AgentSetting } = storeToRefs(settingStore)
const { searchingMode, isFilteringMode } = storeToRefs(echoStore)
const mobileSearchOpen = ref(false)
const activeTab = computed<'home' | 'publish' | 'status' | 'tags' | 'hub'>(() => {
  if (route.query.tab === 'publish' && isLogin.value) return 'publish'
  if (route.query.tab === 'status') return 'status'
  if (route.query.tab === 'tags') return 'tags'
  if (route.query.tab === 'hub') return 'hub'
  return 'home'
})
const shouldHideBannerOnMobile = computed(
  () => mobileSearchOpen.value || searchingMode.value || isFilteringMode.value,
)

const mainColumn = ref<HTMLElement | null>(null)
const TIMELINE_SCROLL_KEY = 'home:timeline:scrollTop'
const WINDOW_SCROLL_KEY = 'home:window:scrollTop'
let timelineScrollRaf: number | null = null
let windowScrollRaf: number | null = null

const paletteOpen = ref<boolean>(false)

const handleGlobalKeydown = (event: KeyboardEvent) => {
  const isSearchShortcut =
    (event.metaKey || event.ctrlKey) && !event.altKey && !event.shiftKey && event.key === 'k'
  if (isSearchShortcut) {
    event.preventDefault()
    paletteOpen.value = !paletteOpen.value
    return
  }
  if (event.key === 'Escape' && paletteOpen.value) {
    paletteOpen.value = false
  }
}

const saveTimelineScrollPosition = () => {
  if (!mainColumn.value || timelineScrollRaf !== null) return
  timelineScrollRaf = window.requestAnimationFrame(() => {
    timelineScrollRaf = null
    if (!mainColumn.value) return
    sessionStorage.setItem(TIMELINE_SCROLL_KEY, String(mainColumn.value.scrollTop))
  })
}

const saveWindowScrollPosition = () => {
  if (windowScrollRaf !== null) return
  windowScrollRaf = window.requestAnimationFrame(() => {
    windowScrollRaf = null
    sessionStorage.setItem(WINDOW_SCROLL_KEY, String(window.scrollY))
  })
}

const restoreTimelineScrollPosition = () => {
  if (mainColumn.value) {
    const raw = sessionStorage.getItem(TIMELINE_SCROLL_KEY)
    const scrollTop = raw ? Number(raw) : 0
    if (Number.isFinite(scrollTop) && scrollTop > 0) {
      mainColumn.value.scrollTop = scrollTop
    }
  }
  const rawWindow = sessionStorage.getItem(WINDOW_SCROLL_KEY)
  const windowTop = rawWindow ? Number(rawWindow) : 0
  if (Number.isFinite(windowTop) && windowTop > 0) {
    window.scrollTo({ top: windowTop, behavior: 'auto' })
  }
}

const prefetchHeavyChunks = () => {
  const trigger = () => {
    import('@/views/echo/EchoView.vue').catch(() => {})
    import('@/editor/core/markdown').catch(() => {})
  }
  const ric = (window as Window & { requestIdleCallback?: typeof requestIdleCallback })
    .requestIdleCallback
  if (typeof ric === 'function') {
    ric(trigger, { timeout: 2000 })
  } else {
    window.setTimeout(trigger, 1500)
  }
}

onMounted(async () => {
  if (mainColumn.value) {
    mainColumn.value.scrollLeft = 0
    mainColumn.value.addEventListener('scroll', saveTimelineScrollPosition, { passive: true })
  }
  window.addEventListener('scroll', saveWindowScrollPosition, { passive: true })
  const stopScrollRestoreWatch = watch(
    () => echoStore.echoList.length > 0 && !echoStore.isLoading,
    (ready) => {
      if (!ready) return
      stopScrollRestoreWatch()
      window.requestAnimationFrame(() => {
        restoreTimelineScrollPosition()
      })
    },
    { immediate: true, flush: 'post' },
  )
  window.addEventListener('keydown', handleGlobalKeydown)
  prefetchHeavyChunks()
})

onBeforeUnmount(() => {
  if (mainColumn.value) {
    mainColumn.value.removeEventListener('scroll', saveTimelineScrollPosition)
  }
  if (timelineScrollRaf !== null) {
    window.cancelAnimationFrame(timelineScrollRaf)
    timelineScrollRaf = null
  }
  if (windowScrollRaf !== null) {
    window.cancelAnimationFrame(windowScrollRaf)
    windowScrollRaf = null
  }
  window.removeEventListener('scroll', saveWindowScrollPosition)
  window.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<style scoped>
.home-page {
  --home-canvas: var(--color-bg-canvas, #ebe6df);
  --home-rail-left: 15rem;
  --home-rail-right: 19rem;
  --home-main-max: 36rem;

  min-height: 100dvh;
  background: var(--home-canvas);
  color: var(--color-text-primary);
}

/* 桌面端：整个 home-page 视口高度，左右栏 sticky，中栏自滚 */
@media (width >= 820px) {
  .home-page {
    height: 100dvh;
    overflow: hidden;
  }
}

/* === 移动端 (<820px)：单栏 === */
.home-shell {
  margin: 0 auto;
  padding: 0 0.75rem;
  max-width: 36rem;
}

@media (width >= 640px) {
  .home-shell {
    padding: 0 1rem;
  }
}

.home-shell__left,
.home-shell__right {
  display: none;
}

.home-shell__mobile-top {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 1rem;
  padding-bottom: 0.5rem;
}

.home-shell__main {
  width: 100%;
  min-width: 0;
  padding-bottom: 2rem;
}

/* === 平板 (820-1099px)：左栏 + 中栏 === */
@media (width >= 820px) {
  .home-shell {
    display: grid;
    grid-template-columns: var(--home-rail-left) minmax(0, 1fr);
    gap: 0;
    max-width: calc(var(--home-rail-left) + var(--home-main-max));
    padding: 0;
    height: 100%;
  }

  .home-shell__left {
    display: flex;
    height: 100%;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-gutter: stable;
    border-right: 1px solid var(--color-border-subtle);
  }

  .home-shell__main {
    height: 100%;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-gutter: stable;
    padding: 1.25rem 1.5rem 2rem;
  }

  .home-shell__main--unclipped {
    overflow: visible auto;
  }

  /* 桌面布局下隐藏中栏顶部移动端品牌 */
  .home-shell__mobile-top {
    display: none;
  }
}

/* === 桌面 (≥1100px)：完整三栏 === */
@media (width >= 1100px) {
  .home-shell {
    grid-template-columns: var(--home-rail-left) minmax(0, 1fr) var(--home-rail-right);
    max-width: calc(var(--home-rail-left) + var(--home-main-max) + var(--home-rail-right));
  }

  .home-shell__right {
    display: flex;
    height: 100%;
    overflow-y: auto;
    overscroll-behavior: contain;
    scrollbar-gutter: stable;
    border-left: 1px solid var(--color-border-subtle);
  }
}

/* === 主内容容器内边距：移动端紧凑 === */
.home-content-block {
  width: 100%;
}

.home-content-block--publish {
  padding-inline: 0;
}

.home-status-widgets {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.home-status-widgets__card {
  padding: 0.5rem !important;
}

@media (width >= 768px) {
  .home-content-block--publish {
    padding-inline: 1.25rem;
  }
}

@media (width >= 1024px) {
  .home-content-block--publish {
    padding-inline: 1.5rem;
  }
}

@media (width <= 819.98px) {
  .home-banner--mobile-hidden {
    display: none;
  }
}
</style>
