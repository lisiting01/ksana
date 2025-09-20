import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useSettingsStore = defineStore('settings', () => {
  const apiBase = ref(localStorage.getItem('apiBase') || 'http://localhost:7100')
  const locale = ref(localStorage.getItem('locale') || 'zh-CN')

  watch(apiBase, (newValue) => {
    localStorage.setItem('apiBase', newValue)
  })

  watch(locale, (newValue) => {
    localStorage.setItem('locale', newValue)
  })

  const updateApiBase = (url: string) => {
    apiBase.value = url
  }

  const updateLocale = (lang: string) => {
    locale.value = lang
  }

  return {
    apiBase,
    locale,
    updateApiBase,
    updateLocale
  }
})