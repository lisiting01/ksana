export function formatUTCTime(isoString?: string): string {
  if (!isoString) return '-'

  try {
    const date = new Date(isoString)
    return date.toISOString()
  } catch {
    return isoString
  }
}

export function getLocalTimeTooltip(isoString?: string): string {
  if (!isoString) return ''

  try {
    const date = new Date(isoString)
    return `本地时间: ${date.toLocaleString()}`
  } catch {
    return ''
  }
}

export function generateISOExample(): string {
  const now = new Date()
  now.setMinutes(now.getMinutes() + 30)
  return now.toISOString()
}

export function validateISODateTime(value: string): boolean {
  try {
    const date = new Date(value)
    return date.toISOString() === value
  } catch {
    return false
  }
}