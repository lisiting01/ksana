import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useSettingsStore = defineStore('settings', () => {
  const apiBase = ref(localStorage.getItem('apiBase') || import.meta.env.VITE_API_BASE_URL || 'http://localhost:7100')
  const locale = ref(localStorage.getItem('locale') || 'zh-CN')
  const apiKey = ref(localStorage.getItem('apiKey') || import.meta.env.VITE_API_KEY || '')

  watch(apiBase, (newValue) => {
    localStorage.setItem('apiBase', newValue)
  })

  watch(locale, (newValue) => {
    localStorage.setItem('locale', newValue)
  })

  watch(apiKey, (newValue) => {
    localStorage.setItem('apiKey', newValue)
  })

  const updateApiBase = (url: string) => {
    apiBase.value = url
  }

  const updateLocale = (lang: string) => {
    locale.value = lang
  }

  const updateApiKey = (key: string) => {
    apiKey.value = key
  }

  const clearApiKey = () => {
    apiKey.value = ''
  }

  return {
    apiBase,
    locale,
    apiKey,
    updateApiBase,
    updateLocale,
    updateApiKey,
    clearApiKey
  }
})