<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ElButton, ElMenuItem } from 'element-plus'
import AppLayout from './framework/AppLayout.vue'
</script>

<template>
  <app-layout>
    <template #header-menu>
      <el-button :tag="RouterLink" to="/dashboard/overview" text>控制台</el-button>
    </template>
    <template #sidebar-menu>
      <el-menu-item index="/dashboard/overview">数据看板</el-menu-item>
      <el-menu-item index="/users">用户管理</el-menu-item>
      <el-menu-item index="/groups">分组管理</el-menu-item>
    </template>

    <router-view />
  </app-layout>
</template>
