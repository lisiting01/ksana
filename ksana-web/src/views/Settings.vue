<template>
  <div class="settings-page">
    <div class="header">
      <div class="title">
        <h1>系统设置</h1>
      </div>
    </div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      style="max-width: 600px"
    >
      <el-card header="API 配置" style="margin-bottom: 20px">
        <el-form-item
          label="API 地址"
          prop="apiBase"
          required
        >
          <el-input
            v-model="form.apiBase"
            placeholder="http://localhost:7100"
          />
          <div class="help-text">
            ksana-service 后端服务的地址，包含协议和端口号
          </div>
        </el-form-item>

        <el-form-item
          label="API 密钥"
          prop="apiKey"
        >
          <el-input
            v-model="form.apiKey"
            :type="showApiKey ? 'text' : 'password'"
            placeholder="请输入API密钥"
            clearable
          >
            <template #append>
              <el-button
                @click="showApiKey = !showApiKey"
                :icon="showApiKey ? 'View' : 'Hide'"
                text
                style="width: 40px;"
              />
            </template>
          </el-input>
          <div class="help-text">
            API访问密钥，用于访问受保护的API接口。密钥将以明文形式存储在浏览器本地存储中。
            <el-button
              type="danger"
              text
              size="small"
              @click="clearApiKey"
              style="margin-left: 8px;"
            >
              清空密钥
            </el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="testConnection" :loading="testing">
            测试连接
          </el-button>
          <el-button @click="testAuth" :loading="testingAuth" v-if="form.apiKey">
            测试鉴权
          </el-button>
          <span v-if="connectionStatus" class="connection-status">
            <el-icon :class="connectionStatus.success ? 'success' : 'error'">
              <component :is="connectionStatus.success ? 'CircleCheck' : 'CircleClose'" />
            </el-icon>
            {{ connectionStatus.message }}
          </span>
        </el-form-item>

        <el-alert
          v-if="authStatus"
          :type="authStatus.success ? 'success' : 'error'"
          :title="authStatus.message"
          show-icon
          :closable="false"
          style="margin-top: 12px;"
        />
      </el-card>

      <el-card header="界面配置" style="margin-bottom: 20px">
        <el-form-item label="语言">
          <el-select v-model="form.locale" placeholder="选择语言">
            <el-option label="简体中文" value="zh-CN" />
            <el-option label="English" value="en-US" />
          </el-select>
          <div class="help-text">
            界面显示语言（当前版本仅支持中文）
          </div>
        </el-form-item>
      </el-card>

      <el-card header="系统信息">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="前端版本">
            ksana-web v0.1.0 (MVP)
          </el-descriptions-item>
          <el-descriptions-item label="构建信息">
            Vue 3 + Vite + Element Plus + TypeScript
          </el-descriptions-item>
          <el-descriptions-item label="当前 API 地址">
            <a :href="settingsStore.apiBase" target="_blank" rel="noopener">
              {{ settingsStore.apiBase }}
            </a>
          </el-descriptions-item>
          <el-descriptions-item label="存储位置">
            localStorage (浏览器本地存储)
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <div class="form-actions">
        <el-button type="primary" @click="saveSettings">
          保存设置
        </el-button>
        <el-button @click="resetSettings">
          重置为默认值
        </el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { FormInstance } from 'element-plus'
import { ElMessage } from 'element-plus'
import { CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { useSettingsStore } from '@/stores/settings'
import { request } from '@/api/http'

const settingsStore = useSettingsStore()
const formRef = ref<FormInstance>()
const testing = ref(false)

const form = reactive({
  apiBase: '',
  locale: '',
  apiKey: ''
})

const showApiKey = ref(false)
const testingAuth = ref(false)

const connectionStatus = ref<{
  success: boolean
  message: string
} | null>(null)

const authStatus = ref<{
  success: boolean
  message: string
} | null>(null)

const rules = {
  apiBase: [
    { required: true, message: '请输入 API 地址', trigger: 'blur' },
    { validator: validateApiBase, trigger: 'blur' }
  ]
}

function validateApiBase(_: any, value: string, callback: Function) {
  if (!value) {
    callback(new Error('请输入 API 地址'))
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

const testConnection = async () => {
  if (!form.apiBase) {
    ElMessage.warning('请先输入 API 地址')
    return
  }

  try {
    testing.value = true
    connectionStatus.value = null
    authStatus.value = null

    const response = await fetch(`${form.apiBase}/health`)
    const data = await response.json()

    if (response.ok && data.status === 'ok') {
      connectionStatus.value = {
        success: true,
        message: '连接成功（/health 端点无需鉴权）'
      }
    } else {
      connectionStatus.value = {
        success: false,
        message: '服务状态异常'
      }
    }
  } catch (error) {
    connectionStatus.value = {
      success: false,
      message: error instanceof Error ? error.message : '连接失败'
    }
  } finally {
    testing.value = false
  }
}

const testAuth = async () => {
  if (!form.apiBase || !form.apiKey) {
    ElMessage.warning('请先输入 API 地址和密钥')
    return
  }

  try {
    testingAuth.value = true
    authStatus.value = null

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Authorization': `ApiKey ${form.apiKey}`,
      'X-API-Key': form.apiKey
    }

    const response = await fetch(`${form.apiBase}/jobs`, {
      method: 'GET',
      headers
    })

    if (response.ok) {
      authStatus.value = {
        success: true,
        message: '鉴权成功，API密钥有效'
      }
    } else if (response.status === 401 || response.status === 403) {
      authStatus.value = {
        success: false,
        message: '鉴权失败，API密钥无效或已过期'
      }
    } else {
      authStatus.value = {
        success: false,
        message: `鉴权测试失败: ${response.status} ${response.statusText}`
      }
    }
  } catch (error) {
    authStatus.value = {
      success: false,
      message: error instanceof Error ? error.message : '鉴权测试失败'
    }
  } finally {
    testingAuth.value = false
  }
}

const clearApiKey = () => {
  form.apiKey = ''
  authStatus.value = null
  ElMessage.success('密钥已清空')
}

const saveSettings = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()

    settingsStore.updateApiBase(form.apiBase)
    settingsStore.updateLocale(form.locale)
    settingsStore.updateApiKey(form.apiKey)

    ElMessage.success('设置保存成功')

    // 如果 API 地址变更，重新测试连接
    if (form.apiBase !== settingsStore.apiBase) {
      setTimeout(testConnection, 500)
    }
  } catch (error) {
    console.error('保存设置失败:', error)
  }
}

const resetSettings = () => {
  form.apiBase = 'http://localhost:7100'
  form.locale = 'zh-CN'
  form.apiKey = ''
  connectionStatus.value = null
  authStatus.value = null
}

const loadSettings = () => {
  form.apiBase = settingsStore.apiBase
  form.locale = settingsStore.locale
  form.apiKey = settingsStore.apiKey
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings-page {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.header {
  margin-bottom: 20px;
}

.title h1 {
  margin: 0;
  font-size: 24px;
  color: #333;
}

.help-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.connection-status {
  margin-left: 12px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
}

.connection-status .success {
  color: #67c23a;
}

.connection-status .error {
  color: #f56c6c;
}

.form-actions {
  text-align: center;
  padding: 20px;
}

.form-actions .el-button {
  margin: 0 10px;
}
</style>