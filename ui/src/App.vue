<script setup>
import { nextTick, onMounted, reactive, ref } from 'vue'
import {
  ElAside, ElAvatar, ElButton, ElCard, ElConfigProvider, ElContainer,
  ElDialog, ElEmpty, ElForm, ElFormItem, ElHeader, ElInput, ElMain,
  ElMenu, ElMenuItem, ElMessage, ElMessageBox, ElRow, ElScrollbar,
  ElSpace, ElTable, ElTableColumn, ElText, vLoading,
} from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { usersApi } from './api/users'

const currentView = ref('console')
const users = ref([])
const busy = ref(false)
const dialogVisible = ref(false)
const editingId = ref(null)
const formRef = ref()
const form = reactive({ name: '', email: '' })
const rules = {
  name: [{ required: true, whitespace: true, message: '请输入姓名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
}

async function loadUsers() {
  busy.value = true
  try {
    users.value = await usersApi.list()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    busy.value = false
  }
}

async function openForm(user) {
  editingId.value = user?.id ?? null
  form.name = user?.name ?? ''
  form.email = user?.email ?? ''
  dialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

async function saveUser() {
  if (busy.value || !formRef.value) return
  form.name = form.name.trim()
  form.email = form.email.trim()
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid || busy.value) return
  busy.value = true
  try {
    if (editingId.value === null) await usersApi.create(form)
    else await usersApi.update(editingId.value, form)
    dialogVisible.value = false
    ElMessage.success('已保存')
    await loadUsers()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    busy.value = false
  }
}

async function deleteUser(user) {
  try {
    await ElMessageBox.confirm(`确定删除用户「${user.name}」吗？`, '删除用户', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消',
    })
  } catch {
    return
  }
  busy.value = true
  try {
    await usersApi.delete(user.id)
    ElMessage.success('已删除')
    await loadUsers()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    busy.value = false
  }
}

onMounted(loadUsers)
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
              :default-active="currentView"
              style="border-bottom: none; background: transparent;"
              @select="currentView = $event"
            >
              <el-menu-item index="console">控制台</el-menu-item>
            </el-menu>
            <el-avatar :size="32">U</el-avatar>
          </el-space>
        </el-row>
      </el-header>

      <el-container style="overflow: hidden;">
        <el-aside width="200px" style="border-right: 1px solid var(--el-border-color-light);">
          <el-scrollbar>
            <el-menu
              :default-active="currentView"
              style="border-right: none;"
              @select="currentView = $event"
            >
              <el-menu-item index="users">用户管理</el-menu-item>
            </el-menu>
          </el-scrollbar>
        </el-aside>

        <el-main>
          <el-empty
            v-if="currentView === 'console'"
            description="控制台暂无数据"
          >
            <el-button type="primary" @click="currentView = 'users'">进入用户管理</el-button>
          </el-empty>

          <template v-else>
            <el-card shadow="never">
              <template #header>
                <el-row justify="space-between" align="middle">
                  <el-text size="large" tag="b">用户管理</el-text>
                  <el-button type="primary" :disabled="busy" @click="openForm()">新增用户</el-button>
                </el-row>
              </template>
              <el-table v-loading="busy" :data="users" row-key="id" border empty-text="暂无用户">
                <el-table-column prop="id" label="ID" width="70" />
                <el-table-column prop="name" label="姓名" min-width="120" show-overflow-tooltip />
                <el-table-column prop="email" label="邮箱" min-width="200" show-overflow-tooltip />
                <el-table-column label="操作" width="130" fixed="right">
                  <template #default="{ row }">
                    <el-button link type="primary" :disabled="busy" @click="openForm(row)">编辑</el-button>
                    <el-button link type="danger" :disabled="busy" @click="deleteUser(row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>

            <el-dialog
              v-model="dialogVisible"
              :title="editingId === null ? '新增用户' : '编辑用户'"
              width="min(420px, calc(100vw - 32px))"
              :close-on-click-modal="!busy"
              :close-on-press-escape="!busy"
              :show-close="!busy"
            >
              <el-form ref="formRef" :model="form" :rules="rules" label-position="top" :disabled="busy" @submit.prevent="saveUser">
                <el-form-item label="姓名" prop="name">
                  <el-input v-model="form.name" placeholder="请输入姓名" :maxlength="100" autocomplete="name" />
                </el-form-item>
                <el-form-item label="邮箱" prop="email">
                  <el-input v-model="form.email" placeholder="请输入邮箱地址" :maxlength="254" autocomplete="email" @keyup.enter="saveUser" />
                </el-form-item>
              </el-form>
              <template #footer>
                <el-button :disabled="busy" @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" :loading="busy" @click="saveUser">保存</el-button>
              </template>
            </el-dialog>
          </template>
        </el-main>
      </el-container>
    </el-container>
  </el-config-provider>
</template>
