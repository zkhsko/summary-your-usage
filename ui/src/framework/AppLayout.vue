<script setup>
import {
  ElAside, ElAvatar, ElConfigProvider, ElContainer, ElHeader, ElMain,
  ElMenu, ElRow, ElScrollbar, ElSpace, ElText,
} from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

defineProps({ activeView: { type: String, required: true } })
defineEmits(['select'])
</script>

<template>
  <el-config-provider :locale="zhCn">
    <el-container style="height: 100vh;">
      <el-header style="border-bottom: 1px solid var(--el-border-color-light);">
        <el-row justify="space-between" align="middle" style="height: 100%;">
          <el-text size="large" tag="b">Summary Your Usage</el-text>
          <el-space size="large">
            <el-menu
              mode="horizontal"
              :ellipsis="false"
              :default-active="activeView"
              style="border-bottom: none; background: transparent;"
              @select="$emit('select', $event)"
            >
              <slot name="header-menu" />
            </el-menu>
            <el-avatar :size="32">U</el-avatar>
          </el-space>
        </el-row>
      </el-header>

      <el-container style="overflow: hidden;">
        <el-aside width="200px" style="border-right: 1px solid var(--el-border-color-light);">
          <el-scrollbar>
            <el-menu
              :default-active="activeView"
              style="border-right: none;"
              @select="$emit('select', $event)"
            >
              <slot name="sidebar-menu" />
            </el-menu>
          </el-scrollbar>
        </el-aside>

        <el-main>
          <slot />
        </el-main>
      </el-container>
    </el-container>
  </el-config-provider>
</template>
