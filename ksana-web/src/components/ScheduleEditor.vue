<template>
  <div class="schedule-editor">
    <el-form-item label="调度类型" required>
      <el-radio-group v-model="localSchedule.kind" @change="onKindChange">
        <el-radio value="once">一次性任务</el-radio>
        <el-radio value="every">周期性任务</el-radio>
      </el-radio-group>
    </el-form-item>

    <template v-if="localSchedule.kind === 'once'">
      <el-form-item
        label="运行时间"
        required
        :rules="[
          { required: true, message: '请输入运行时间' },
          { validator: validateRunAt, trigger: 'blur' }
        ]"
        prop="schedule.run_at"
      >
        <el-input
          v-model="(localSchedule as ScheduleOnce).run_at"
          placeholder="请输入 UTC 时间，格式：2025-09-20T03:00:00Z"
        />
        <div class="help-text">
          使用 UTC 时间，格式为 RFC3339。
          <el-button type="text" size="small" @click="generateExample">
            生成示例时间
          </el-button>
        </div>
      </el-form-item>
    </template>

    <template v-else-if="localSchedule.kind === 'every'">
      <el-form-item
        label="执行间隔"
        required
        :rules="[
          { required: true, message: '请输入执行间隔' },
          { validator: validateDuration, trigger: 'blur' }
        ]"
        prop="schedule.every"
      >
        <el-input
          v-model="(localSchedule as ScheduleEvery).every"
          placeholder="例如：5m、1h、30s"
        />
        <div class="help-text">
          Go 时间格式：s(秒)、m(分钟)、h(小时)。例如：{{ getDurationExamples().join('、') }}
        </div>
      </el-form-item>

      <el-form-item
        label="开始时间"
        :rules="[{ validator: validateStartAt, trigger: 'blur' }]"
      >
        <el-input
          v-model="(localSchedule as ScheduleEvery).start_at"
          placeholder="可选，留空表示立即开始"
        />
        <div class="help-text">
          UTC 时间，格式为 RFC3339。留空表示立即开始执行。
        </div>
      </el-form-item>

      <el-form-item
        label="随机延迟"
        :rules="[{ validator: validateJitter, trigger: 'blur' }]"
      >
        <el-input
          v-model="(localSchedule as ScheduleEvery).jitter"
          placeholder="可选，例如：30s"
        />
        <div class="help-text">
          可选的随机延迟时间，用于避免多个任务同时执行。
        </div>
      </el-form-item>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Schedule, ScheduleOnce, ScheduleEvery } from '@/types/job'
import { validateDuration as validateDurationUtil, getDurationExamples } from '@/utils/schedule'
import { generateISOExample, validateISODateTime } from '@/utils/time'

interface Props {
  modelValue: Schedule
}

interface Emits {
  (e: 'update:modelValue', value: Schedule): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const localSchedule = ref<Schedule>({ ...props.modelValue })

watch(
  () => props.modelValue,
  (newValue) => {
    localSchedule.value = { ...newValue }
  },
  { deep: true }
)

watch(
  localSchedule,
  (newValue) => {
    emit('update:modelValue', newValue)
  },
  { deep: true }
)

const onKindChange = () => {
  if (localSchedule.value.kind === 'once') {
    localSchedule.value = {
      kind: 'once',
      run_at: generateISOExample()
    } as ScheduleOnce
  } else {
    localSchedule.value = {
      kind: 'every',
      every: '5m',
      start_at: null,
      jitter: null
    } as ScheduleEvery
  }
}

const generateExample = () => {
  if (localSchedule.value.kind === 'once') {
    ;(localSchedule.value as ScheduleOnce).run_at = generateISOExample()
  }
}

const validateRunAt = (_: any, value: string, callback: Function) => {
  if (!value) {
    callback(new Error('请输入运行时间'))
    return
  }
  if (!validateISODateTime(value)) {
    callback(new Error('请输入正确的 UTC 时间格式'))
    return
  }
  callback()
}

const validateDuration = (_: any, value: string, callback: Function) => {
  if (!value) {
    callback(new Error('请输入执行间隔'))
    return
  }
  if (!validateDurationUtil(value)) {
    callback(new Error('请输入正确的时间格式，例如：5m、1h、30s'))
    return
  }
  callback()
}

const validateStartAt = (_: any, value: string, callback: Function) => {
  if (value && !validateISODateTime(value)) {
    callback(new Error('请输入正确的 UTC 时间格式'))
    return
  }
  callback()
}

const validateJitter = (_: any, value: string, callback: Function) => {
  if (value && !validateDurationUtil(value)) {
    callback(new Error('请输入正确的时间格式，例如：30s'))
    return
  }
  callback()
}
</script>

<style scoped>
.schedule-editor {
  margin-bottom: 20px;
}

.help-text {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}
</style>