import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(ElementPlus)

// 确保设置store在应用启动时初始化
import { useSettingsStore } from './stores/settings'
const settingsStore = useSettingsStore()

// 从localStorage恢复设置
// 这在store定义中已经自动处理了，但这里可以进行额外的初始化逻辑

app.mount('#app')
