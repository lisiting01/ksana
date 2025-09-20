<template>
  <div class="job-form">
    <div class="header">
      <div class="title">
        <h1>{{ isEdit ? '编辑任务' : '新建任务' }}</h1>
      </div>
      <div class="actions">
        <el-button @click="$router.back()">返回</el-button>
      </div>
    </div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      style="max-width: 800px"
    >
      <el-card header="基础信息" style="margin-bottom: 20px">
        <el-form-item
          label="任务名称"
          required
          prop="name"
        >
          <el-input
            v-model="form.name"
            placeholder="请输入任务名称"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="启用状态">
          <el-switch
            v-model="form.enabled"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>

        <el-form-item label="任务类型">
          <el-select v-model="form.type" disabled>
            <el-option label="HTTP 回调" value="http" />
          </el-select>
          <div class="help-text">当前版本仅支持 HTTP 回调类型</div>
        </el-form-item>
      </el-card>

      <el-card header="HTTP 配置" style="margin-bottom: 20px">
        <HttpConfigEditor v-model="form.http" />
      </el-card>

      <el-card header="调度配置" style="margin-bottom: 20px">
        <ScheduleEditor v-model="form.schedule" />
      </el-card>

      <el-card header="执行控制" style="margin-bottom: 20px">
        <el-form-item
          label="超时时间"
          prop="timeout"
          :rules="[{ validator: validateTimeout, trigger: 'blur' }]"
        >
          <el-input
            v-model="form.timeout"
            placeholder="例如：10s、1m"
            style="width: 200px"
          />
          <div class="help-text">
            HTTP 请求的超时时间，格式如：{{ getDurationExamples().join('、') }}
          </div>
        </el-form-item>

        <el-form-item
          label="最大重试次数"
          prop="max_retries"
        >
          <el-input-number
            v-model="form.max_retries"
            :min="0"
            :max="10"
            style="width: 200px"
          />
          <div class="help-text">失败时的最大重试次数</div>
        </el-form-item>

        <el-form-item
          label="重试间隔"
          prop="retry_backoff"
          :rules="[{ validator: validateRetryBackoff, trigger: 'blur' }]"
        >
          <el-input
            v-model="form.retry_backoff"
            placeholder="例如：5s、30s"
            style="width: 200px"
          />
          <div class="help-text">每次重试之间的等待时间</div>
        </el-form-item>
      </el-card>

      <div class="form-actions">
        <el-button
          type="primary"
          :loading="loading"
          @click="submitForm"
        >
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
        <el-button @click="resetForm">重置</el-button>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { FormInstance } from 'element-plus'
import { useJobsStore } from '@/stores/jobs'
import { validateDuration as validateDurationUtil, getDurationExamples } from '@/utils/schedule'
import { generateISOExample } from '@/utils/time'
import type { CreateJobRequest, UpdateJobRequest, JobResponse } from '@/types/job'
import ScheduleEditor from '@/components/ScheduleEditor.vue'
import HttpConfigEditor from '@/components/HttpConfigEditor.vue'

const route = useRoute()
const router = useRouter()
const jobsStore = useJobsStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const isEdit = computed(() => route.name === 'job-edit')
const jobId = computed(() => route.params.id as string)

const form = reactive<CreateJobRequest & { id?: string }>({
  name: '',
  enabled: true,
  type: 'http',
  http: {
    method: 'GET',
    url: '',
    headers: {},
    body: ''
  },
  schedule: {
    kind: 'every',
    every: '5m',
    start_at: null,
    jitter: null
  },
  timeout: '10s',
  max_retries: 3,
  retry_backoff: '5s'
})

const rules = {
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' },
    { min: 1, max: 100, message: '任务名称长度在 1 到 100 个字符', trigger: 'blur' }
  ]
}

const validateTimeout = (_: any, value: string, callback: Function) => {
  if (value && !validateDurationUtil(value)) {
    callback(new Error('请输入正确的时间格式'))
    return
  }
  callback()
}

const validateRetryBackoff = (_: any, value: string, callback: Function) => {
  if (value && !validateDurationUtil(value)) {
    callback(new Error('请输入正确的时间格式'))
    return
  }
  callback()
}

const loadJob = async () => {
  if (!isEdit.value) return

  try {
    loading.value = true
    const job = await jobsStore.getById(jobId.value)?.value
    if (!job) {
      await jobsStore.fetchAll()
      const updatedJob = jobsStore.getById(jobId.value)?.value
      if (!updatedJob) {
        throw new Error('任务不存在')
      }
      Object.assign(form, updatedJob)
    } else {
      Object.assign(form, job)
    }
  } catch (error) {
    console.error('加载任务失败:', error)
    router.push('/jobs')
  } finally {
    loading.value = false
  }
}

const submitForm = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    // 处理空字符串字段，将其转换为 null
    const processSchedule = (schedule: any) => {
      const processed = { ...schedule }
      if (processed.kind === 'every') {
        if (processed.start_at === '') processed.start_at = null
        if (processed.jitter === '') processed.jitter = null
      }
      return processed
    }

    if (isEdit.value) {
      const updateData: UpdateJobRequest = {
        name: form.name,
        enabled: form.enabled,
        http: form.http,
        schedule: processSchedule(form.schedule),
        timeout: form.timeout,
        max_retries: form.max_retries,
        retry_backoff: form.retry_backoff
      }
      await jobsStore.update(jobId.value, updateData)
    } else {
      const createData: CreateJobRequest = {
        name: form.name,
        enabled: form.enabled,
        type: 'http',
        http: form.http,
        schedule: processSchedule(form.schedule),
        timeout: form.timeout,
        max_retries: form.max_retries,
        retry_backoff: form.retry_backoff
      }
      await jobsStore.create(createData)
    }

    router.push('/jobs')
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  if (isEdit.value) {
    loadJob()
  } else {
    Object.assign(form, {
      name: '',
      enabled: true,
      type: 'http',
      http: {
        method: 'GET',
        url: '',
        headers: {},
        body: ''
      },
      schedule: {
        kind: 'every',
        every: '5m',
        start_at: null,
        jitter: null
      },
      timeout: '10s',
      max_retries: 3,
      retry_backoff: '5s'
    })
  }
  formRef.value?.resetFields()
}

onMounted(() => {
  if (isEdit.value) {
    loadJob()
  }
})
</script>

<style scoped>
.job-form {
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

.help-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.form-actions {
  text-align: center;
  padding: 20px;
}

.form-actions .el-button {
  margin: 0 10px;
}
</style>