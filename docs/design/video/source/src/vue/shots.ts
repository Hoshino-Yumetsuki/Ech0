// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Registry of mountable product compositions: real Ech0 views, mounted into film host elements.
import type { Component } from 'vue'
import HomeView from '@/views/home/HomeView.vue'
import ChatPage from '@/views/chat/modules/ChatPage.vue'
import PanelView from '@/views/panel/PanelView.vue'
import EchoView from '@/views/echo/EchoView.vue'

export const SHOTS: Record<string, Component> = {
  home: HomeView,
  chat: ChatPage,
  panel: PanelView,
  echo: EchoView,
}
