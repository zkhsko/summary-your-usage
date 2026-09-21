<script setup>
import { ref } from 'vue'
import { ElButton, ElMenuItem } from 'element-plus'
import AppLayout from './framework/AppLayout.vue'
import OverviewPage from './features/dashboard/OverviewPage.vue'
import UsersPage from './features/users/UsersPage.vue'

const currentView = ref('dashboard/overview')
</script>

<template>
  <app-layout :active-view="currentView" @select="currentView = $event">
    <template #header-menu>
      <el-button text @click="currentView = 'dashboard/overview'">控制台</el-button>
    </template>
    <template #sidebar-menu>
      <el-menu-item index="users">用户管理</el-menu-item>
    </template>

    <overview-page v-if="currentView === 'dashboard/overview'" />
    <users-page v-else-if="currentView === 'users'" />
  </app-layout>
</template>
