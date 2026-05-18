<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <aside class="home-right-rail">
    <!-- 搜索 + 标签/日期 chip -->
    <div class="home-right-rail__filter">
      <TheFilter @open-palette="emit('openPalette')" />
    </div>

    <!-- 常驻 widget：每个 widget 包在独立 surface card 里，与中栏 echo 卡片视觉一致 -->
    <PanelCard border-style="solid" class="home-right-rail__widget">
      <TheHeatMap />
    </PanelCard>
    <PanelCard border-style="solid" class="home-right-rail__widget">
      <TheConnectWidget />
    </PanelCard>
    <PanelCard border-style="solid" class="home-right-rail__widget">
      <TheCommentWidget />
    </PanelCard>
    <PanelCard v-if="AgentSetting.enable" border-style="solid" class="home-right-rail__widget">
      <TheRecentCard />
    </PanelCard>
    <PanelCard border-style="solid" class="home-right-rail__widget">
      <TheTagPileWidget />
    </PanelCard>

    <!-- 底部 version + footer -->
    <div class="home-right-rail__footer">
      <a
        href="https://github.com/lin-snow/Ech0"
        target="_blank"
        rel="noopener noreferrer"
        class="home-right-rail__version"
      >
        version: {{ settingStore.hello?.version || '--' }}
      </a>
      <template v-if="footerContent">
        <a
          v-if="footerLink"
          :href="footerLink"
          target="_blank"
          rel="noopener noreferrer"
          class="home-right-rail__footer-link"
        >
          {{ footerContent }}
        </a>
        <span v-else class="home-right-rail__footer-link">{{ footerContent }}</span>
      </template>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { storeToRefs } from 'pinia'
import { useSettingStore } from '@/stores'
import TheFilter from './TheFilter.vue'
import PanelCard from '@/layout/PanelCard.vue'
import {
  TheCommentWidget,
  TheConnectWidget,
  TheHeatMap,
  TheRecentCard,
} from '@/components/advanced/widget'

// TagPile 较重，懒加载
const TheTagPileWidget = defineAsyncComponent(
  () => import('@/components/advanced/widget/TheTagPileWidget.vue'),
)

const emit = defineEmits<{
  (e: 'openPalette'): void
}>()

const settingStore = useSettingStore()
const { SystemSetting, AgentSetting } = storeToRefs(settingStore)
const footerContent = computed(
  () => SystemSetting.value.footer_content || SystemSetting.value.ICP_number,
)
const footerLink = computed(() => SystemSetting.value.footer_link)
</script>

<style scoped>
.home-right-rail {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  padding: 1.25rem 0.5rem 1.5rem;
  min-width: 0;
}

.home-right-rail__filter {
  /* filter 自身已带 search-shell 样式，这里只保留留白容器 */
  padding: 0.25rem 0.25rem 0;
}

.home-right-rail__widget {
  /* widget 内部已经有自己的 head 和 padding，这里把 PanelCard 的内边距压低 */
  padding: 0.5rem !important;
}

.home-right-rail__footer {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.25rem 0.75rem 0;
  margin-top: auto;
}

.home-right-rail__version,
.home-right-rail__footer-link {
  display: inline-block;
  font-family: var(--font-family-display);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-decoration: none;
  letter-spacing: 0.02em;
  transition: color 0.15s ease;
}

.home-right-rail__version:hover,
.home-right-rail__footer-link:hover {
  color: var(--color-text-primary);
}
</style>
