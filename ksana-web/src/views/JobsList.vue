<template>
  <div class="jobs-list">
    <div class="header">
      <div class="title">
        <h1>任务列表</h1>
      </div>
      <div class="actions">
        <el-button type="primary" @click="$router.push('/jobs/new')">
          新建任务
        </el-button>
        <el-button @click="refresh">刷新</el-button>
      </div>
    </div>

    <div class="filters">
      <el-form :inline="true">
        <el-form-item label="名称">
          <el-input
            v-model="filters.name"
            placeholder="搜索任务名称"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="filters.enabled"
            placeholder="选择状态"
            clearable
            style="width: 120px"
          >
            <el-option label="全部" :value="undefined" />
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select
            v-model="filters.type"
            placeholder="选择类型"
            clearable
            style="width: 120px"
          >
            <el-option label="全部" :value="undefined" />
            <el-option label="HTTP" value="http" />
          </el-select>
        </el-form-item>
      </el-form>
    </div>

    <el-table
      v-loading="jobsStore.loading"
      :data="filteredJobs"
      stripe
      border
    >
      <el-table-column prop="name" label="任务名称" min-width="150" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'danger'">
            {{ row.enabled ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="80">
        <template #default="{ row }">
          <el-tag type="info">{{ row.type.toUpperCase() }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="调度配置" min-width="200">
        <template #default="{ row }">
          {{ getScheduleSummary(row.schedule) }}
        </template>
      </el-table-column>
      <el-table-column label="下次运行" width="180">
        <template #default="{ row }">
          <span :title="getLocalTimeTooltip(row.next_run_at)">
            {{ formatUTCTime(row.next_run_at) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="最后状态" width="100">
        <template #default="{ row }">
          <el-tag
            v-if="row.last_status"
            :type="getStatusType(row.last_status)"
            size="small"
          >
            {{ getStatusText(row.last_status) }}
          </el-tag>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="最后运行" width="180">
        <template #default="{ row }">
          <span :title="getLocalTimeTooltip(row.last_run_at)">
            {{ formatUTCTime(row.last_run_at) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewJob(row.id)">查看</el-button>
          <el-button size="small" @click="editJob(row.id)">编辑</el-button>
          <el-dropdown
            @command="(command: string) => handleAction(command, row)"
            trigger="click"
          >
            <el-button size="small">
              更多 <el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="run-now">立即执行</el-dropdown-item>
                <el-dropdown-item :command="row.enabled ? 'pause' : 'resume'">
                  {{ row.enabled ? '暂停' : '恢复' }}
                </el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
      <template #empty>
        <div class="empty-state">
          <el-empty description="暂无任务数据">
            <el-button type="primary" @click="$router.push('/jobs/new')">
              创建第一个任务
            </el-button>
          </el-empty>
        </div>
      </template>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { useJobsStore } from '@/stores/jobs'
import { getScheduleSummary } from '@/utils/schedule'
import { formatUTCTime, getLocalTimeTooltip } from '@/utils/time'
import { handleAuthError } from '@/utils/auth'
import type { JobResponse } from '@/types/job'

const router = useRouter()
const jobsStore = useJobsStore()

const filters = ref({
  name: '',
  enabled: undefined as boolean | undefined,
  type: undefined as string | undefined
})

const filteredJobs = computed(() => {
  const jobs = jobsStore.items || []
  return jobs.filter(job => {
    if (filters.value.name && !job.name.toLowerCase().includes(filters.value.name.toLowerCase())) {
      return false
    }
    if (filters.value.enabled !== undefined && job.enabled !== filters.value.enabled) {
      return false
    }
    if (filters.value.type && job.type !== filters.value.type) {
      return false
    }
    return true
  })
})

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

const refresh = async () => {
  try {
    await jobsStore.fetchAll()
  } catch (error) {
    if (!handleAuthError(error, router)) {
      console.error('获取任务列表失败:', error)
    }
  }
}

const viewJob = (id: string) => {
  router.push(`/jobs/${id}`)
}

const editJob = (id: string) => {
  router.push(`/jobs/${id}/edit`)
}

const handleAction = async (command: string, job: JobResponse) => {
  try {
    switch (command) {
      case 'run-now':
        await jobsStore.runNow(job.id)
        await refresh()
        break
      case 'pause':
        await jobsStore.pause(job.id)
        break
      case 'resume':
        await jobsStore.resume(job.id)
        break
      case 'delete':
        await ElMessageBox.confirm(
          `确定要删除任务 "${job.name}" 吗？`,
          '确认删除',
          {
            confirmButtonText: '删除',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        await jobsStore.remove(job.id)
        break
    }
  } catch (error) {
    if (!handleAuthError(error, router)) {
      console.error('操作失败:', error)
    }
  }
}

onMounted(() => {
  refresh()
})
</script>

<style scoped>
.jobs-list {
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

.filters {
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
}

.empty-state {
  padding: 20px 0;
}
</style>