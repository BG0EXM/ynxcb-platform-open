<template>
  <el-container class="layout">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <div class="logo-badge">
          <img src="../assets/danghui.png" alt="党徽" class="danghui-img" />
        </div>
        <div class="logo-text">
          <h2>伊宁县委宣传部</h2>
          <p>部务工作平台</p>
        </div>
      </div>
      <el-menu :default-active="$route.path" router background-color="#1f2d3d"
        text-color="#a7bfd9" active-text-color="#ffffff">
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-badge :value="unread" :hidden="!unread" class="unread-badge">
            <el-icon class="header-icon" @click="$router.push('/incoming')"><Bell /></el-icon>
          </el-badge>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="30" class="user-avatar">{{ avatarText }}</el-avatar>
              <span class="user-name">{{ authStore.user?.real_name }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
      <el-footer class="app-footer" height="40px">
        <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener">ICP备案号占位</a>
        <span class="footer-sep">|</span>
        <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener">
          公网安备号占位
        </a>
      </el-footer>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import request from '../utils/request'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const menus = computed(() => {
  const all = [
    { path: '/dashboard', title: '工作台', icon: 'Odometer' },
    { path: '/incoming', title: '收文管理', icon: 'FolderOpened' },
    { path: '/study', title: '公共资料', icon: 'Reading' },
    { path: '/contacts', title: '通讯录', icon: 'Phone' },
    { path: '/duty', title: '值守排班', icon: 'AlarmClock' },
    { path: '/reports', title: '周/月/年报', icon: 'Tickets' },
    { path: '/attendance', title: '考勤点到', icon: 'AlarmClock' },
    { path: '/leave', title: '请假管理', icon: 'Calendar' },
    { path: '/vehicles', title: '公车管理', icon: 'Van' }
  ]
  if (authStore.isAdmin) {
    all.push({ path: '/users', title: '用户管理', icon: 'User' })
  }
  return all
})

const currentTitle = computed(() => route.meta.title || '')

const unread = ref(0)

const loadUnread = async () => {
  try {
    const res = await request.get('/dashboard-stats')
    unread.value = res.pending_incoming || 0
  } catch (e) {}
}

const refreshUnread = () => {
  loadUnread()
}

onMounted(() => {
  loadUnread()
  // 事件驱动刷新：收文操作后触发
  window.addEventListener('incoming-changed', refreshUnread)
})
onUnmounted(() => {
  window.removeEventListener('incoming-changed', refreshUnread)
})

const avatarText = computed(() => {
  const name = authStore.user?.real_name || '用户'
  return name.charAt(name.length - 1)
})

const handleCommand = (cmd) => {
  if (cmd === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (cmd === 'profile') {
    router.push('/profile')
  }
}
</script>

<style scoped>
.layout {
  height: 100%;
}
.sidebar {
  background: #1f2d3d;
  overflow-x: hidden;
}
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fff;
  border-bottom: 1px solid rgba(255,255,255,0.1);
  padding: 0 8px;
}
.logo-badge {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}
.danghui-img {
  width: 34px;
  height: 34px;
  object-fit: contain;
}
.logo-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
}
.logo h2 {
  font-size: 15px;
  color: #fff;
  margin-bottom: 2px;
  line-height: 1.2;
  white-space: nowrap;
}
.logo p {
  font-size: 10px;
  color: #a7bfd9;
  letter-spacing: 2px;
  line-height: 1.2;
}
.sidebar :deep(.el-menu) {
  border-right: none;
}
.header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}
.header-icon {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  margin-right: 20px;
}
.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  gap: 8px;
}
.user-avatar {
  background: #c8102e;
}
.user-name {
  color: #303133;
  font-size: 14px;
}
.main {
  background: #f0f2f5;
  padding: 16px;
}
.unread-badge {
  margin-right: 8px;
}
.app-footer {
  background: #fff;
  border-top: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 12px;
  color: #909399;
}
.app-footer a {
  color: #909399;
  text-decoration: none;
}
.app-footer a:hover {
  color: #409eff;
}
.footer-sep {
  color: #c0c4cc;
}
</style>
