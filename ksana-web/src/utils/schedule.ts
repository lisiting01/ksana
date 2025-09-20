import type { Schedule } from '@/types/job'

export function getScheduleSummary(schedule: Schedule): string {
  if (schedule.kind === 'once') {
    return `once ${schedule.run_at}`
  } else {
    const parts = [`every ${schedule.every}`]
    if (schedule.start_at) {
      parts.push(`(start ${schedule.start_at})`)
    } else {
      parts.push('(start now)')
    }
    if (schedule.jitter) {
      parts.push(`jitter ${schedule.jitter}`)
    }
    return parts.join(' ')
  }
}

export function validateDuration(duration: string): boolean {
  if (!duration) return false
  const durationRegex = /^\d+[hms](\d+[hms])*$/
  return durationRegex.test(duration)
}

export function getDefaultDuration(): string {
  return '10s'
}

export function getDurationExamples(): string[] {
  return ['5s', '1m', '1h', '30m', '2h30m']
}