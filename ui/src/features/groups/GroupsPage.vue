<script setup>
import { nextTick, onMounted, reactive, ref } from 'vue'
import {
  ElButton, ElDialog, ElForm, ElFormItem, ElInput, ElInputNumber, ElMessage,
  ElMessageBox, ElRow, ElSpace, ElSwitch, ElTable, ElTableColumn, ElTag, ElText, vLoading,
} from 'element-plus'
import { groupsApi } from './api'

const groups = ref([])
const busy = ref(false)
const dialogVisible = ref(false)
const editingId = ref(null)
const formRef = ref()
const form = reactive({ name: '', billing_multiplier: 1, visible_other_group: false, description: '' })
const rules = {
  name: [{ required: true, whitespace: true, message: '请输入分组名', trigger: 'blur' }],
  billing_multiplier: [
    { required: true, type: 'number', min: 0, message: '请输入有效的非负计费倍率', trigger: 'change' },
  ],
}

async function loadGroups() {
  busy.value = true
  try {
    groups.value = await groupsApi.list()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    busy.value = false
  }
}

async function openForm(group) {
  editingId.value = group?.id ?? null
  form.name = group?.name ?? ''
  form.billing_multiplier = group?.billing_multiplier ?? 1
  form.visible_other_group = group?.visible_other_group ?? false
  form.description = group?.description ?? ''
  dialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

async function saveGroup() {
  if (busy.value || !formRef.value) return
  form.name = form.name.trim()
  form.description = form.description.trim()
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid || busy.value) return
  busy.value = true
  try {
    if (editingId.value === null) await groupsApi.create(form)
    else await groupsApi.update(editingId.value, form)
    dialogVisible.value = false
    ElMessage.success('已保存')
    await loadGroups()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    busy.value = false
  }
}

async function deleteGroup(group) {
  if (busy.value) return
  try {
    await ElMessageBox.confirm(`确定删除分组「${group.name}」吗？`, '删除分组', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消',
    })
  } catch {
    return
  }
  if (busy.value) return
  busy.value = true
  try {
    await groupsApi.delete(group.id)
    ElMessage.success('已删除')
    await loadGroups()
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    busy.value = false
  }
}

onMounted(loadGroups)
</script>

<template>
  <el-space direction="vertical" fill :size="24" style="width: 100%;">
    <el-row justify="space-between" align="middle">
      <el-text size="large" tag="b">分组管理</el-text>
      <el-button type="primary" :disabled="busy" @click="openForm()">新增分组</el-button>
    </el-row>
    <el-table v-loading="busy" :data="groups" row-key="id" empty-text="暂无分组">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="分组名" min-width="140" show-overflow-tooltip />
      <el-table-column prop="billing_multiplier" label="计费倍率" width="130" />
      <el-table-column label="非当前组用户可见" width="170">
        <template #default="{ row }">
          <el-tag :type="row.visible_other_group ? 'success' : 'info'">
            {{ row.visible_other_group ? '可见' : '不可见' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明信息" min-width="200" show-overflow-tooltip />
      <el-table-column label="操作" width="130" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" :disabled="busy" @click="openForm(row)">编辑</el-button>
          <el-button link type="danger" :disabled="busy" @click="deleteGroup(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-space>

  <el-dialog
    v-model="dialogVisible"
    :title="editingId === null ? '新增分组' : '编辑分组'"
    width="min(520px, calc(100vw - 32px))"
    :close-on-click-modal="!busy"
    :close-on-press-escape="!busy"
    :show-close="!busy"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" :disabled="busy" @submit.prevent="saveGroup">
      <el-form-item label="分组名" prop="name">
        <el-input v-model="form.name" placeholder="请输入唯一的分组名" :maxlength="100" />
      </el-form-item>
      <el-form-item label="计费倍率" prop="billing_multiplier">
        <el-input-number v-model="form.billing_multiplier" :min="0" :step="0.1" controls-position="right" style="width: 100%;" />
        <el-text size="small" type="info">1 为标准倍率，0 为免费</el-text>
      </el-form-item>
      <el-form-item label="非当前组用户可见" prop="visible_other_group">
        <el-switch v-model="form.visible_other_group" active-text="可见" inactive-text="不可见" />
      </el-form-item>
      <el-form-item label="说明信息" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="4" :maxlength="1000" show-word-limit placeholder="请输入分组说明（选填）" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="busy" @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="busy" @click="saveGroup">保存</el-button>
    </template>
  </el-dialog>
</template>
