import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/http'

export function handleAuthError(error: unknown, router?: ReturnType<typeof useRouter>) {
  if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
    ElMessage.error(error.message)

    ElMessageBox.confirm(
      '请在设置页面配置有效的API密钥后重试',
      '需要API密钥',
      {
        confirmButtonText: '前往设置',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      if (router) {
        router.push('/settings')
      } else {
        // 如果没有传入router，使用全局路由
        const currentRouter = useRouter()
        currentRouter.push('/settings')
      }
    }).catch(() => {
      // 用户取消，不做任何操作
    })

    return true
  }
  return false
}