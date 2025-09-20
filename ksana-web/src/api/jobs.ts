import type { CreateJobRequest, UpdateJobRequest, JobResponse, HealthResponse, ActionResponse } from '@/types/job'
import { request } from './http'

export const JobsAPI = {
  list: () => request<JobResponse[]>('/jobs'),

  get: (id: string) => request<JobResponse>(`/jobs/${id}`),

  create: (payload: CreateJobRequest) =>
    request<JobResponse>('/jobs', {
      method: 'POST',
      body: JSON.stringify(payload)
    }),

  update: (id: string, payload: UpdateJobRequest) =>
    request<JobResponse>(`/jobs/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(payload)
    }),

  remove: (id: string) =>
    request<void>(`/jobs/${id}`, { method: 'DELETE' }),

  runNow: (id: string) =>
    request<ActionResponse>(`/jobs/${id}/run-now`, { method: 'POST' }),

  pause: (id: string) =>
    request<ActionResponse>(`/jobs/${id}/pause`, { method: 'POST' }),

  resume: (id: string) =>
    request<ActionResponse>(`/jobs/${id}/resume`, { method: 'POST' }),

  health: () => request<HealthResponse>('/health'),
}