import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

request.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  response => response.data,
  error => {
    const status = error.response?.status
    const msg = error.response?.data?.error || error.message
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (router.currentRoute.value.path !== '/login') {
        router.push('/login')
      }
    }
    ElMessage.error(msg || '请求失败')
    return Promise.reject(error)
  }
)

// 下载导出文件（Excel），触发浏览器下载
export async function exportFile(url, params = {}) {
  const token = localStorage.getItem('token')
  try {
    // 用全局 axios 获取 blob（不走 JSON 响应拦截器）
    const res = await axios.get(url, {
      baseURL: '/api',
      params,
      responseType: 'blob',
      headers: { Authorization: `Bearer ${token}` },
      timeout: 60000
    })
    // 检查是否为 JSON 错误响应（如 401 时后端返回 {"error":...}）
    if (res.data && res.data.type === 'application/json') {
      const text = await res.data.text()
      const err = JSON.parse(text)
      ElMessage.error(err.error || '导出失败')
      return
    }
    // 从 Content-Disposition 提取文件名
    let fileName = '导出.xlsx'
    const cd = res.headers['content-disposition']
    if (cd) {
      const match = cd.match(/filename\*=UTF-8''([^;]+)/)
      if (match && match[1]) {
        try {
          fileName = decodeURIComponent(match[1])
        } catch (e) {
          fileName = '导出.xlsx'
        }
      }
    }
    // 用 FileReader 兼容方式触发下载（更稳）
    const blob = new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    setTimeout(() => URL.revokeObjectURL(link.href), 1000)
    ElMessage.success('导出成功')
  } catch (e) {
    // 401 时跳登录，其他错误提示
    if (e.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
    } else {
      ElMessage.error('导出失败，请重试')
    }
  }
}

export default request
