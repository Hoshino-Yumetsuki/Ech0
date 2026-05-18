<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <article class="echo-card group w-full">
    <header class="echo-card__header flex justify-between items-center">
      <div class="flex justify-start items-center h-9 gap-1.5">
        <span class="echo-card__dot" aria-hidden="true">
          <span class="w-2 h-2 rounded-full bg-[var(--color-accent)]"></span>
        </span>
        <button
          type="button"
          @click="handleExpandEcho(echo.id)"
          class="flex items-center h-full justify-start leading-none text-sm font-semibold text-nowrap text-[var(--color-accent)] cursor-pointer hover:underline hover:decoration-offset-3 hover:decoration-1"
        >
          {{ formatDate(props.echo.created_at) }}
        </button>
        <button
          type="button"
          class="echo-open-btn flex items-center justify-center w-6 h-6 rounded-sm text-[var(--color-text-muted)] opacity-0 group-hover:opacity-100 transition-opacity duration-150 hover:text-[var(--color-text-primary)] focus-visible:opacity-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--color-border-subtle)]"
          :aria-label="t('echoCard.openDetail')"
          v-tooltip="t('echoCard.openDetail')"
          @click="handleExpandEcho(echo.id)"
        >
          <Open class="w-3.5 h-3.5" />
        </button>
      </div>

      <div
        v-if="!userStore.isLogin && props.echo.private"
        v-tooltip="t('echoCard.privateStatus')"
        class="w-7 h-7 flex items-center justify-center bg-[var(--color-bg-muted)] ring-1 ring-[var(--color-border-subtle)] ring-inset rounded-full"
      >
        <Lock class="w-4 h-4" />
      </div>

      <div v-else-if="userStore.isLogin" class="relative flex items-center justify-center">
        <button
          ref="menuTriggerRef"
          type="button"
          :aria-label="t('echoCard.moreActions')"
          :aria-expanded="isMenuOpen"
          aria-haspopup="menu"
          class="w-7 h-7 flex items-center justify-center bg-[var(--color-bg-muted)] ring-1 ring-[var(--color-border-subtle)] ring-inset rounded-full transition-shadow duration-150 hover:shadow-sm focus:outline-none"
          @click.stop="toggleMenu"
        >
          <More class="w-5 h-5" />
        </button>

        <Teleport to="body">
          <transition
            enter-active-class="transition duration-150 ease-out"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition duration-100 ease-in"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95"
          >
            <div
              v-if="isMenuOpen"
              ref="menuPanelRef"
              :style="menuPanelStyle"
              class="fixed z-5000 w-36 origin-top-right rounded-[var(--radius-md)] bg-[var(--color-bg-surface)] ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-md)] p-1"
            >
              <div
                v-if="props.echo.private"
                class="flex items-center gap-1.5 px-2 pt-1 pb-1.5 text-[var(--color-text-muted)]"
              >
                <Lock class="w-3 h-3" />
                <span class="text-[10px] font-semibold tracking-[0.08em] uppercase">
                  {{ t('echoCard.privateStatus') }}
                </span>
              </div>

              <button
                type="button"
                class="menu-row"
                @click="
                  () => {
                    closeMenu()
                    handleUpdateEcho()
                  }
                "
              >
                <EditEcho class="w-3.5 h-3.5 shrink-0" />
                <span>{{ t('echoCard.update') }}</span>
              </button>
              <button
                type="button"
                class="menu-row menu-row--danger"
                @click="
                  () => {
                    closeMenu()
                    handleDeleteEcho(props.echo.id)
                  }
                "
              >
                <Roll class="w-3.5 h-3.5 shrink-0" />
                <span>{{ t('echoCard.delete') }}</span>
              </button>
            </div>
          </transition>
        </Teleport>
      </div>
    </header>

    <div class="echo-card__body">
      <template
        v-if="
          props.echo.layout === ImageLayout.GRID ||
          props.echo.layout === ImageLayout.HORIZONTAL ||
          props.echo.layout === ImageLayout.STACK
        "
      >
        <div class="echo-card__content">
          <TheMdPreview :content="props.echo.content" />
        </div>

        <TheImageGallery
          :images="echoImageFiles"
          :layout="props.echo.layout"
          :priority="props.index === 0"
        />
      </template>

      <template v-else>
        <TheImageGallery
          :images="echoImageFiles"
          :layout="props.echo.layout"
          :priority="props.index === 0"
        />

        <div class="echo-card__content echo-card__content--after-gallery">
          <TheMdPreview :content="props.echo.content" />
        </div>
      </template>

      <div v-if="props.echo.extension" class="echo-card__extension">
        <TheExtensionRenderer :echo="props.echo" />
      </div>
    </div>
  </article>
</template>

<script lang="ts">
import { ref } from 'vue'

// Module-scoped: whichever echo id is currently showing its action menu.
// Declared in a non-setup <script> so it runs once per module, not per instance.
// Setting this id auto-closes any other card's menu (only one open at a time).
const activeMenuId = ref<string | null>(null)
</script>

<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import { fetchDeleteEcho, fetchGetEchoById } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useUserStore, useEchoStore, useEditorStore } from '@/stores'
import { TheMdPreview } from '@/components/advanced/md'
import Roll from '@/components/icons/roll.vue'
import Lock from '@/components/icons/lock.vue'
import More from '@/components/icons/more.vue'
import EditEcho from '@/components/icons/editecho.vue'
import Open from '@/components/icons/open.vue'
import { useRouter } from 'vue-router'
import { ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy } from '@/utils/echo'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useI18n } from 'vue-i18n'

const TheImageGallery = defineAsyncComponent(
  () => import('@/components/advanced/gallery/TheImageGallery.vue'),
)
const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

const { openConfirm } = useBaseDialog()
const { t } = useI18n()

const emit = defineEmits(['refresh'])

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
  index?: number
}>()

const userStore = useUserStore()
const echoImageFiles = computed(() =>
  getEchoFilesBy(props.echo, { categories: ['image'], dedupeBy: 'id' }),
)

const echoStore = useEchoStore()
const editorStore = useEditorStore()
const router = useRouter()

const handleDeleteEcho = (echoId: string) => {
  openConfirm({
    title: String(t('echoCard.deleteConfirmTitle')),
    description: String(t('echoCard.deleteConfirmDesc')),
    onConfirm: () => {
      fetchDeleteEcho(echoId).then(() => {
        theToast.success(String(t('echoCard.deleteSuccess')))
        emit('refresh')
      })
    },
  })
}

const handleUpdateEcho = async () => {
  if (editorStore.isUpdateMode) {
    window.scrollTo({ top: 0, behavior: 'smooth' })
    theToast.warning(String(t('echoCard.exitUpdateModeFirst')))
    return
  }

  const res = await fetchGetEchoById(String(props.echo.id))
  if (res.code === 1 && res.data) {
    echoStore.echoToUpdate = res.data
  } else {
    echoStore.echoToUpdate = props.echo
  }

  editorStore.isUpdateMode = true
  await router.push({
    name: 'home',
    query: { tab: 'publish' },
  })
}

const handleExpandEcho = (echoId: string) => {
  router.push({
    name: 'echo',
    params: { echoId: echoId },
  })
}

const isMenuOpen = computed(() => activeMenuId.value === props.echo.id)
const menuTriggerRef = ref<HTMLElement | null>(null)
const menuPanelRef = ref<HTMLElement | null>(null)
const menuPanelStyle = ref<Record<string, string>>({})

const MENU_GAP = 8

const updateMenuPosition = () => {
  if (!isMenuOpen.value || !menuTriggerRef.value) return
  const rect = menuTriggerRef.value.getBoundingClientRect()
  const viewportRight = window.innerWidth
  const right = Math.max(MENU_GAP, viewportRight - rect.right)
  menuPanelStyle.value = {
    top: `${rect.bottom + MENU_GAP}px`,
    right: `${right}px`,
  }
}

const closeMenu = () => {
  if (isMenuOpen.value) activeMenuId.value = null
}

const toggleMenu = () => {
  activeMenuId.value = isMenuOpen.value ? null : props.echo.id
}

watch(isMenuOpen, async (open) => {
  if (open) {
    await nextTick()
    updateMenuPosition()
  }
})

const handleDocumentClick = (event: MouseEvent) => {
  if (!isMenuOpen.value) return
  const target = event.target as Node | null
  if (!target) return
  if (menuPanelRef.value?.contains(target) || menuTriggerRef.value?.contains(target)) return
  closeMenu()
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape') closeMenu()
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  document.addEventListener('keydown', handleEscape)
  window.addEventListener('resize', updateMenuPosition)
  window.addEventListener('scroll', updateMenuPosition, true)
})

onBeforeUnmount(() => {
  if (isMenuOpen.value) activeMenuId.value = null
  document.removeEventListener('click', handleDocumentClick)
  document.removeEventListener('keydown', handleEscape)
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
})
</script>

<style scoped lang="css">
/* 离散卡片：bg-surface 在加深的 canvas 上明显"浮起"，
   保留日期点 marker 作为内容平台的视觉锚点 */
.echo-card {
  position: relative;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  padding: 0.5rem 0.75rem 0.75rem;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.echo-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-soft);
}

.echo-card__header {
  position: relative;
  z-index: 1;
}

.echo-card__dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 0.625rem;
  height: 100%;
  flex-shrink: 0;
}

.echo-card__body {
  margin-top: 0.25rem;
}

.echo-card__content {
  padding: 0 0.25rem;
  margin-bottom: 0.5rem;
}

.echo-card__content--after-gallery {
  margin-top: 0.75rem;
  margin-bottom: 0;
}

.echo-card__extension {
  margin-top: 0.5rem;
}

.echo-card:hover .echo-open-btn {
  animation: echo-open-nudge-left 1200ms ease-out both;
}

.echo-open-btn:focus-visible {
  animation: echo-open-nudge-left 1200ms ease-out both;
}

@keyframes echo-open-nudge-left {
  0% {
    transform: translateX(0);
  }

  16% {
    transform: translateX(-4px);
  }

  28% {
    transform: translateX(1px);
  }

  42% {
    transform: translateX(-3px);
  }

  54% {
    transform: translateX(1px);
  }

  68% {
    transform: translateX(-2px);
  }

  80% {
    transform: translateX(0);
  }

  100% {
    transform: translateX(0);
  }
}

.menu-row {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 8px;
  height: 28px;
  padding: 0 8px;
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 500;
  text-align: left;
  transition:
    color 150ms ease,
    background-color 150ms ease;
}

.menu-row:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.menu-row:focus-visible {
  outline: none;
  background: var(--color-bg-muted);
  box-shadow: inset 0 0 0 1.5px var(--color-border-strong);
}

.menu-row--danger:hover {
  color: var(--color-danger);
}
</style>
