// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { localStg } from '@/utils/storage'

export type ThemeMode = 'light' | 'dark' | 'sunny' | 'system'
type ThemeType = 'light' | 'dark' | 'sunny'
const THEME_COLOR_META_NAME = 'theme-color'
const THEME_COLOR_FALLBACK: Record<ThemeType, string> = {
  light: '#f4f1ec',
  dark: '#333333',
  sunny: '#eeece6',
}

const SYSTEM_DARK_QUERY = '(prefers-color-scheme: dark)'

const isThemeMode = (value: unknown): value is ThemeMode =>
  value === 'light' || value === 'dark' || value === 'sunny' || value === 'system'

const isThemeType = (value: unknown): value is ThemeType =>
  value === 'light' || value === 'dark' || value === 'sunny'

// 读取系统当前的颜色偏好；服务端或不支持 matchMedia 的环境下默认返回 'light'
const getSystemTheme = (): ThemeType => {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return 'light'
  }
  return window.matchMedia(SYSTEM_DARK_QUERY).matches ? 'dark' : 'light'
}

export const useThemeStore = defineStore('themeStore', () => {
  const savedThemeMode = localStg.getItem('themeMode')
  const savedTheme = localStg.getItem('theme')

  // 用户选择的主题模式：未存储或非法值时默认跟随系统
  const mode = ref<ThemeMode>(isThemeMode(savedThemeMode) ? savedThemeMode : 'system')
  // 实际生效的主题：只可能是 light / dark / sunny 三者之一
  const theme = ref<ThemeType>(isThemeType(savedTheme) ? savedTheme : getSystemTheme())

  // 内部切换主题逻辑：system -> light -> sunny -> dark -> system
  const applyThemeToggle = () => {
    if (mode.value === 'system') {
      mode.value = 'light'
    } else if (mode.value === 'light') {
      mode.value = 'sunny'
    } else if (mode.value === 'sunny') {
      mode.value = 'dark'
    } else {
      mode.value = 'system'
    }

    applyTheme()
    localStg.setItem('themeMode', mode.value)
  }

  // 防抖标志：防止动画过程中重复触发
  let isTransitioning = false

  // 使用 View Transitions 默认交叉淡化（比 clip-path 圆扩散更省 GPU）
  const toggleTheme = async () => {
    if (isTransitioning) return

    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    type ViewTransitionLike = {
      ready: Promise<void>
      finished: Promise<void>
      updateCallbackDone?: Promise<void>
    }
    const startViewTransition = (
      document as Document & {
        startViewTransition?: (callback: () => void) => ViewTransitionLike
      }
    ).startViewTransition?.bind(document)

    if (prefersReducedMotion || !startViewTransition) {
      applyThemeToggle()
      return
    }

    isTransitioning = true
    try {
      const transition = startViewTransition(() => {
        applyThemeToggle()
      })
      await transition.finished
    } finally {
      isTransitioning = false
    }
  }

  const applyTheme = () => {
    // 跟随系统：忽略晴日，仅在浅色 / 深色之间切换
    if (mode.value === 'system') {
      theme.value = getSystemTheme()
    } else {
      theme.value = mode.value
    }

    document.documentElement.classList.remove('light', 'dark', 'sunny')
    document.documentElement.classList.add(theme.value)
    syncThemeColorMeta()
    localStg.setItem('theme', theme.value)
  }

  const syncThemeColorMeta = () => {
    const rootStyles = getComputedStyle(document.documentElement)
    const chromeColor = rootStyles.getPropertyValue('--color-chrome-theme').trim()
    const canvasColor = rootStyles.getPropertyValue('--color-bg-canvas').trim()
    const nextThemeColor = chromeColor || canvasColor || THEME_COLOR_FALLBACK[theme.value]

    let themeColorMeta = document.querySelector<HTMLMetaElement>(
      `meta[name="${THEME_COLOR_META_NAME}"]`,
    )
    if (!themeColorMeta) {
      themeColorMeta = document.createElement('meta')
      themeColorMeta.setAttribute('name', THEME_COLOR_META_NAME)
      document.head.appendChild(themeColorMeta)
    }

    themeColorMeta.setAttribute('content', nextThemeColor)
  }

  // 监听系统配色偏好变化：仅在 mode === 'system' 时同步生效，避免覆盖用户的手动选择
  let systemMediaQuery: MediaQueryList | null = null
  const handleSystemThemeChange = () => {
    if (mode.value === 'system') {
      applyTheme()
    }
  }

  const init = () => {
    applyTheme()

    if (
      !systemMediaQuery &&
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function'
    ) {
      systemMediaQuery = window.matchMedia(SYSTEM_DARK_QUERY)
      systemMediaQuery.addEventListener('change', handleSystemThemeChange)
    }
  }

  return {
    theme,
    mode,
    toggleTheme,
    applyTheme,
    init,
  }
})
