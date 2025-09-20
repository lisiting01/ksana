<template>
  <div class="http-config-editor">
    <el-form-item
      label="请求方法"
      required
      prop="http.method"
      :rules="[{ required: true, message: '请选择请求方法' }]"
    >
      <el-select v-model="localConfig.method" placeholder="选择请求方法">
        <el-option label="GET" value="GET" />
        <el-option label="POST" value="POST" />
      </el-select>
    </el-form-item>

    <el-form-item
      label="请求地址"
      required
      prop="http.url"
      :rules="[
        { required: true, message: '请输入请求地址' },
        { validator: validateUrl, trigger: 'blur' }
      ]"
    >
      <el-input
        v-model="localConfig.url"
        placeholder="https://example.com/api/webhook"
      />
    </el-form-item>

    <el-form-item label="请求头">
      <div class="headers-editor">
        <div
          v-for="(header, index) in headersList"
          :key="index"
          class="header-row"
        >
          <el-input
            v-model="header.key"
            placeholder="Header名称"
            style="width: 200px"
            @input="updateHeaders"
          />
          <el-input
            v-model="header.value"
            placeholder="Header值"
            style="width: 300px; margin-left: 8px"
            @input="updateHeaders"
          />
          <el-button
            type="danger"
            text
            @click="removeHeader(index)"
            style="margin-left: 8px"
          >
            删除
          </el-button>
        </div>
        <el-button type="primary" text @click="addHeader">
          + 添加请求头
        </el-button>
      </div>
    </el-form-item>

    <el-form-item label="请求体" v-if="localConfig.method === 'POST'">
      <el-input
        v-model="localConfig.body"
        type="textarea"
        :rows="6"
        placeholder="请求体内容，通常为 JSON 格式"
      />
      <div class="help-text">
        对于 POST 请求，可以在这里输入请求体内容。如果是 JSON 格式，请确保语法正确。
      </div>
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { HTTPConfig } from '@/types/job'

interface Props {
  modelValue: HTTPConfig
}

interface Emits {
  (e: 'update:modelValue', value: HTTPConfig): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const localConfig = ref<HTTPConfig>({ ...props.modelValue })

interface HeaderItem {
  key: string
  value: string
}

const headersList = ref<HeaderItem[]>([])

const updateHeadersList = () => {
  headersList.value = Object.entries(localConfig.value.headers).map(([key, value]) => ({
    key,
    value
  }))
  if (headersList.value.length === 0) {
    headersList.value.push({ key: '', value: '' })
  }
}

const updateHeaders = () => {
  const headers: Record<string, string> = {}
  headersList.value
    .filter(header => header.key.trim())
    .forEach(header => {
      headers[header.key.trim()] = header.value
    })
  localConfig.value.headers = headers
}

const addHeader = () => {
  headersList.value.push({ key: '', value: '' })
}

const removeHeader = (index: number) => {
  headersList.value.splice(index, 1)
  if (headersList.value.length === 0) {
    headersList.value.push({ key: '', value: '' })
  }
  updateHeaders()
}

watch(
  () => props.modelValue,
  (newValue) => {
    localConfig.value = { ...newValue }
    updateHeadersList()
  },
  { deep: true, immediate: true }
)

watch(
  localConfig,
  (newValue) => {
    emit('update:modelValue', newValue)
  },
  { deep: true }
)

const validateUrl = (_: any, value: string, callback: Function) => {
  if (!value) {
    callback(new Error('请输入请求地址'))
    return
  }
  try {
    const url = new URL(value)
    if (!['http:', 'https:'].includes(url.protocol)) {
      callback(new Error('请输入有效的 HTTP 或 HTTPS 地址'))
      return
    }
  } catch {
    callback(new Error('请输入有效的 URL 地址'))
    return
  }
  callback()
}
</script>

<style scoped>
.http-config-editor {
  margin-bottom: 20px;
}

.headers-editor {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 12px;
  background: #fafafa;
}

.header-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.header-row:last-of-type {
  margin-bottom: 12px;
}

.help-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}
</style>