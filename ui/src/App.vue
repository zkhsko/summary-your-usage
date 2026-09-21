<script setup>
import { ref } from 'vue'
import { ElButton, ElEmpty, ElMenuItem } from 'element-plus'
import AppLayout from './framework/AppLayout.vue'
import UsersPage from './features/users/UsersPage.vue'

const currentView = ref('console')
</script>

<template>
  <app-layout :active-view="currentView" @select="currentView = $event">
    <template #header-menu>
      <el-menu-item index="console">控制台</el-menu-item>
    </template>
    <template #sidebar-menu>
      <el-menu-item index="users">用户管理</el-menu-item>
    </template>

    <el-empty v-if="currentView === 'console'" description="控制台暂无数据">
      <el-button type="primary" @click="currentView = 'users'">进入用户管理</el-button>
    </el-empty>
    <users-page v-else-if="currentView === 'users'" />
  </app-layout>
</template>
