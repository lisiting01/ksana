export type UUID = string

export interface DurationString extends String {}

export interface HTTPConfig {
  method: 'GET' | 'POST'
  url: string
  headers: Record<string, string>
  body: string
}

export type ScheduleKind = 'once' | 'every'

export interface ScheduleOnce {
  kind: 'once'
  run_at: string
}

export interface ScheduleEvery {
  kind: 'every'
  every: DurationString
  start_at?: string | null
  jitter?: DurationString | null
}

export type Schedule = ScheduleOnce | ScheduleEvery

export interface CreateJobRequest {
  name: string
  enabled?: boolean
  type: 'http'
  http: HTTPConfig
  schedule: Schedule
  timeout?: DurationString
  max_retries?: number
  retry_backoff?: DurationString
}

export interface UpdateJobRequest {
  name?: string
  enabled?: boolean
  http?: HTTPConfig
  schedule?: Schedule
  timeout?: DurationString
  max_retries?: number
  retry_backoff?: DurationString
}

export interface JobResponse {
  id: UUID
  name: string
  enabled: boolean
  type: 'http'
  http: HTTPConfig
  schedule: Schedule
  timeout: DurationString
  max_retries: number
  retry_backoff: DurationString
  last_run_at?: string
  next_run_at?: string
  last_status: 'success' | 'failed' | 'timeout' | 'skipped' | 'paused' | 'missed' | ''
  last_error: string
}

export interface HealthResponse {
  status: string
}

export interface ActionResponse {
  message: string
}