<template>
  <div class="health-page">
    <div class="header">
      <div class="title">
        <h1>健康检查</h1>
      </div>
      <div class="actions">
        <el-button @click="checkHealth">刷新</el-button>
      </div>
    </div>

    <el-card>
      <div v-loading="loading" class="health-content">
        <div v-if="healthData" class="health-status">
          <div class="status-indicator">
            <el-icon
              :class="['status-icon', healthData.status === 'ok' ? 'healthy' : 'unhealthy']"
              :size="48"
            >
              <component :is="healthData.status === 'ok' ? 'CircleCheck' : 'CircleClose'" />
            </el-icon>
            <div class="status-text">
              <h2>{{ healthData.status === 'ok' ? '服务正常' : '服务异常' }}</h2>
              <p class="status-description">
                {{ healthData.status === 'ok' ? 'ksana-service 运行正常' : 'ksana-service 运行异常' }}
              </p>
            </div>
          </div>

          <el-divider />

          <div class="health-details">
            <el-descriptions title="服务信息" :column="2" border>
              <el-descriptions-item label="状态">
                <el-tag :type="healthData.status === 'ok' ? 'success' : 'danger'">
                  {{ healthData.status }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="检查时间">
                {{ formatTime(lastCheckTime) }}
              </el-descriptions-item>
              <el-descriptions-item label="API 地址">
                <a :href="apiBase" target="_blank" rel="noopener">
                  {{ apiBase }}
                </a>
              </el-descriptions-item>
              <el-descriptions-item label="健康检查端点">
                <a :href="`${apiBase}/health`" target="_blank" rel="noopener">
                  {{ apiBase }}/health
                </a>
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <el-divider />

          <div class="quick-actions">
            <h3>快速操作</h3>
            <div class="action-buttons">
              <el-button type="primary" @click="$router.push('/jobs')">
                查看任务列表
              </el-button>
              <el-button @click="$router.push('/jobs/new')">
                创建新任务
              </el-button>
              <el-button @click="$router.push('/settings')">
                系统设置
              </el-button>
            </div>
          </div>
        </div>

        <div v-else-if="error" class="error-state">
          <el-result
            icon="error"
            title="无法连接到服务"
            :sub-title="error"
          >
            <template #extra>
              <el-button type="primary" @click="checkHealth">
                重新检查
              </el-button>
              <el-button @click="$router.push('/settings')">
                检查设置
              </el-button>
            </template>
          </el-result>
        </div>
      </div>
    </el-card>

    <div class="tips">
      <el-alert
        title="提示"
        type="info"
        :closable="false"
        show-icon
      >
        <template #default>
          <p>健康检查用于确认 ksana-service 后端服务是否正常运行。</p>
          <ul>
            <li>如果显示"服务正常"，说明前端可以正常与后端通信</li>
            <li>如果显示"服务异常"或连接失败，请检查：
              <ul>
                <li>后端服务是否启动（默认端口 7100）</li>
                <li>API 地址配置是否正确</li>
                <li>网络连接是否正常</li>
              </ul>
            </li>
          </ul>
        </template>
      </el-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { JobsAPI } from '@/api/jobs'
import { apiBase } from '@/api/http'
import type { HealthResponse } from '@/types/job'

const loading = ref(false)
const healthData = ref<HealthResponse>()
const error = ref<string>()
const lastCheckTime = ref<Date>()

const checkHealth = async () => {
  try {
    loading.value = true
    error.value = undefined
    healthData.value = await JobsAPI.health()
    lastCheckTime.value = new Date()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '连接失败'
    healthData.value = undefined
  } finally {
    loading.value = false
  }
}

const formatTime = (time?: Date) => {
  if (!time) return '-'
  return time.toLocaleString()
}

onMounted(() => {
  checkHealth()
})
</script>

<style scoped>
.health-page {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title h1 {
  margin: 0;
  font-size: 24px;
  color: #333;
}

.health-content {
  min-height: 300px;
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  padding: 40px;
  text-align: center;
}

.status-icon {
  margin-bottom: 16px;
}

.status-icon.healthy {
  color: #67c23a;
}

.status-icon.unhealthy {
  color: #f56c6c;
}

.status-text h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: #333;
}

.status-description {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.health-details {
  margin: 20px 0;
}

.quick-actions h3 {
  margin: 0 0 16px 0;
  color: #333;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.error-state {
  padding: 40px;
  text-align: center;
}

.tips {
  margin-top: 20px;
}

.tips ul {
  margin: 8px 0;
  padding-left: 20px;
}

.tips li {
  margin: 4px 0;
}
</style>