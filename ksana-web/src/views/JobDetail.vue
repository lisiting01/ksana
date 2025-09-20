<template>
  <div class="job-detail">
    <div class="header">
      <div class="title">
        <h1>任务详情</h1>
      </div>
      <div class="actions">
        <el-button @click="$router.back()">返回</el-button>
        <el-button @click="editJob">编辑</el-button>
        <el-button type="primary" @click="runNow">立即执行</el-button>
        <el-dropdown @command="handleAction" trigger="click">
          <el-button>
            更多 <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :command="job?.enabled ? 'pause' : 'resume'">
                {{ job?.enabled ? '暂停' : '恢复' }}
              </el-dropdown-item>
              <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div v-if="loading" v-loading="true" class="loading-container"></div>

    <div v-else-if="job" class="content">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-card header="基础信息" style="margin-bottom: 20px">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="任务名称">
                {{ job.name }}
              </el-descriptions-item>
              <el-descriptions-item label="任务 ID">
                <el-tag type="info" size="small">{{ job.id }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="job.enabled ? 'success' : 'danger'">
                  {{ job.enabled ? '启用' : '禁用' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="类型">
                <el-tag type="info">{{ job.type.toUpperCase() }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card header="HTTP 配置" style="margin-bottom: 20px">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="请求方法">
                <el-tag :type="job.http.method === 'GET' ? 'success' : 'primary'">
                  {{ job.http.method }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="请求地址">
                <a :href="job.http.url" target="_blank" rel="noopener">
                  {{ job.http.url }}
                </a>
              </el-descriptions-item>
              <el-descriptions-item label="请求头" v-if="Object.keys(job.http.headers).length">
                <div class="headers-display">
                  <div
                    v-for="(value, key) in job.http.headers"
                    :key="key"
                    class="header-item"
                  >
                    <strong>{{ key }}:</strong> {{ value }}
                  </div>
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="请求体" v-if="job.http.body">
                <pre class="body-content">{{ job.http.body }}</pre>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card header="调度配置" style="margin-bottom: 20px">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="调度类型">
                <el-tag :type="job.schedule.kind === 'once' ? 'warning' : 'primary'">
                  {{ job.schedule.kind === 'once' ? '一次性' : '周期性' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="调度描述">
                {{ getScheduleSummary(job.schedule) }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="job.schedule.kind === 'once'"
                label="运行时间"
              >
                <span :title="getLocalTimeTooltip(job.schedule.run_at)">
                  {{ formatUTCTime(job.schedule.run_at) }}
                </span>
              </el-descriptions-item>
              <el-descriptions-item
                v-if="job.schedule.kind === 'every'"
                label="执行间隔"
              >
                {{ job.schedule.every }}
              </el-descriptions-item>
              <el-descriptions-item
                v-if="job.schedule.kind === 'every' && job.schedule.start_at"
                label="开始时间"
              >
                <span :title="getLocalTimeTooltip(job.schedule.start_at)">
                  {{ formatUTCTime(job.schedule.start_at) }}
                </span>
              </el-descriptions-item>
              <el-descriptions-item
                v-if="job.schedule.kind === 'every' && job.schedule.jitter"
                label="随机延迟"
              >
                {{ job.schedule.jitter }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card header="执行控制" style="margin-bottom: 20px">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="超时时间">
                {{ job.timeout }}
              </el-descriptions-item>
              <el-descriptions-item label="最大重试次数">
                {{ job.max_retries }}
              </el-descriptions-item>
              <el-descriptions-item label="重试间隔">
                {{ job.retry_backoff }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card header="执行状态" style="margin-bottom: 20px">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="最后状态">
                <el-tag
                  v-if="job.last_status"
                  :type="getStatusType(job.last_status)"
                >
                  {{ getStatusText(job.last_status) }}
                </el-tag>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="最后运行">
                <span
                  v-if="job.last_run_at"
                  :title="getLocalTimeTooltip(job.last_run_at)"
                >
                  {{ formatUTCTime(job.last_run_at) }}
                </span>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="下次运行">
                <span
                  v-if="job.next_run_at"
                  :title="getLocalTimeTooltip(job.next_run_at)"
                >
                  {{ formatUTCTime(job.next_run_at) }}
                </span>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="错误信息" v-if="job.last_error">
                <div class="error-message">
                  {{ job.last_error }}
                </div>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card header="JSON 视图">
            <el-button
              @click="showJson = !showJson"
              size="small"
              style="margin-bottom: 10px"
            >
              {{ showJson ? '隐藏' : '显示' }} JSON
            </el-button>
            <pre v-if="showJson" class="json-view">{{ JSON.stringify(job, null, 2) }}</pre>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div v-else class="error-state">
      <el-result
        icon="error"
        title="任务不存在"
        sub-title="请检查任务 ID 是否正确"
      >
        <template #extra>
          <el-button type="primary" @click="$router.push('/jobs')">
            返回列表
          </el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { useJobsStore } from '@/stores/jobs'
import { getScheduleSummary } from '@/utils/schedule'
import { formatUTCTime, getLocalTimeTooltip } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const jobsStore = useJobsStore()

const loading = ref(false)
const showJson = ref(false)

const jobId = computed(() => route.params.id as string)
const job = computed(() => jobsStore.getById(jobId.value)?.value)

const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    success: 'success',
    failed: 'danger',
    timeout: 'warning',
    skipped: 'info',
    paused: 'warning',
    missed: 'danger'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    success: '成功',
    failed: '失败',
    timeout: '超时',
    skipped: '跳过',
    paused: '暂停',
    missed: '错过'
  }
  return statusMap[status] || status
}

const loadJob = async () => {
  if (job.value) return

  try {
    loading.value = true
    await jobsStore.fetchAll()
  } catch (error) {
    console.error('加载任务失败:', error)
  } finally {
    loading.value = false
  }
}

const editJob = () => {
  router.push(`/jobs/${jobId.value}/edit`)
}

const runNow = async () => {
  if (!job.value) return

  try {
    await jobsStore.runNow(job.value.id)
    await loadJob()
  } catch (error) {
    console.error('执行任务失败:', error)
  }
}

const handleAction = async (command: string) => {
  if (!job.value) return

  switch (command) {
    case 'pause':
      await jobsStore.pause(job.value.id)
      break
    case 'resume':
      await jobsStore.resume(job.value.id)
      break
    case 'delete':
      await ElMessageBox.confirm(
        `确定要删除任务 "${job.value.name}" 吗？`,
        '确认删除',
        {
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
      await jobsStore.remove(job.value.id)
      router.push('/jobs')
      break
  }
}

onMounted(() => {
  loadJob()
})
</script>

<style scoped>
.job-detail {
  padding: 20px;
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

.loading-container {
  height: 400px;
}

.headers-display {
  max-height: 200px;
  overflow-y: auto;
}

.header-item {
  margin-bottom: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.body-content {
  max-height: 200px;
  overflow-y: auto;
  background: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.error-message {
  color: #f56c6c;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  max-height: 100px;
  overflow-y: auto;
  background: #fef0f0;
  padding: 8px;
  border-radius: 4px;
}

.json-view {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  margin: 0;
  max-height: 400px;
  overflow-y: auto;
}

.error-state {
  padding: 40px;
  text-align: center;
}
</style>