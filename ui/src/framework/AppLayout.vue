<script setup>
import { useRoute } from 'vue-router'
import {
  ElAside, ElAvatar, ElConfigProvider, ElContainer, ElHeader, ElMain,
  ElMenu, ElRow, ElScrollbar, ElSpace, ElText,
} from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

const route = useRoute()
</script>

<template>
  <el-config-provider :locale="zhCn">
    <el-container style="height: 100vh;">
      <el-header style="border-bottom: 1px solid var(--el-border-color-light);">
        <el-row justify="space-between" align="middle" style="height: 100%;">
          <el-text size="large" tag="b">Summary Your Usage</el-text>
          <el-space size="large">
            <nav aria-label="主导航">
              <slot name="header-menu" />
            </nav>
            <el-avatar :size="32">U</el-avatar>
          </el-space>
        </el-row>
      </el-header>

      <el-container style="overflow: hidden;">
        <el-aside width="200px" style="border-right: 1px solid var(--el-border-color-light);">
          <el-scrollbar>
            <el-menu
              :default-active="route.path"
              router
              style="border-right: none;"
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
