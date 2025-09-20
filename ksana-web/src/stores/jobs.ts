import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { JobResponse, CreateJobRequest, UpdateJobRequest } from '@/types/job'
import { JobsAPI } from '@/api/jobs'
import { ElMessage } from 'element-plus'

export const useJobsStore = defineStore('jobs', () => {
  const items = ref<JobResponse[]>([])
  const loading = ref(false)
  const error = ref<string>()

  const getById = (id: string) => {
    return computed(() => items.value.find(job => job.id === id))
  }

  const fetchAll = async () => {
    try {
      loading.value = true
      error.value = undefined
      items.value = await JobsAPI.list()
    } catch (err) {
      const message = err instanceof Error ? err.message : '获取任务列表失败'
      error.value = message
      ElMessage.error(message)
      throw err
    } finally {
      loading.value = false
    }
  }

  const create = async (payload: CreateJobRequest) => {
    try {
      loading.value = true
      const newJob = await JobsAPI.create(payload)
      items.value.push(newJob)
      ElMessage.success('任务创建成功')
      return newJob
    } catch (err) {
      const message = err instanceof Error ? err.message : '创建任务失败'
      ElMessage.error(message)
      throw err
    } finally {
      loading.value = false
    }
  }

  const update = async (id: string, payload: UpdateJobRequest) => {
    try {
      loading.value = true
      const updatedJob = await JobsAPI.update(id, payload)
      const index = items.value.findIndex(job => job.id === id)
      if (index !== -1) {
        items.value[index] = updatedJob
      }
      ElMessage.success('任务更新成功')
      return updatedJob
    } catch (err) {
      const message = err instanceof Error ? err.message : '更新任务失败'
      ElMessage.error(message)
      throw err
    } finally {
      loading.value = false
    }
  }

  const remove = async (id: string) => {
    try {
      loading.value = true
      await JobsAPI.remove(id)
      const index = items.value.findIndex(job => job.id === id)
      if (index !== -1) {
        items.value.splice(index, 1)
      }
      ElMessage.success('任务删除成功')
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除任务失败'
      ElMessage.error(message)
      throw err
    } finally {
      loading.value = false
    }
  }

  const runNow = async (id: string) => {
    try {
      const response = await JobsAPI.runNow(id)
      ElMessage.success(response.message || '任务已触发执行')
    } catch (err) {
      const message = err instanceof Error ? err.message : '触发任务失败'
      ElMessage.error(message)
      throw err
    }
  }

  const pause = async (id: string) => {
    try {
      const response = await JobsAPI.pause(id)
      const job = items.value.find(j => j.id === id)
      if (job) {
        job.enabled = false
      }
      ElMessage.success(response.message || '任务已暂停')
    } catch (err) {
      const message = err instanceof Error ? err.message : '暂停任务失败'
      ElMessage.error(message)
      throw err
    }
  }

  const resume = async (id: string) => {
    try {
      const response = await JobsAPI.resume(id)
      const job = items.value.find(j => j.id === id)
      if (job) {
        job.enabled = true
      }
      ElMessage.success(response.message || '任务已恢复')
    } catch (err) {
      const message = err instanceof Error ? err.message : '恢复任务失败'
      ElMessage.error(message)
      throw err
    }
  }

  return {
    items,
    loading,
    error,
    getById,
    fetchAll,
    create,
    update,
    remove,
    runNow,
    pause,
    resume
  }
})