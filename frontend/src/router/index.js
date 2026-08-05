import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/incoming/print/:id',
    name: 'incoming-print',
    component: () => import('../views/IncomingPrint.vue'),
    meta: { title: '打印呈批单' }
  },
  {
    path: '/incoming/print-card/:id',
    name: 'incoming-print-card',
    component: () => import('../views/IncomingPrintCard.vue'),
    meta: { title: '打印传阅登记卡' }
  },
  {
    path: '/incoming/label/:id',
    name: 'incoming-label',
    component: () => import('../views/IncomingLabel.vue'),
    meta: { title: '打印标签' }
  },
  {
    path: '/vehicle/print/:id',
    name: 'vehicle-print',
    component: () => import('../views/VehiclePrint.vue'),
    meta: { title: '打印派车单' }
  },
  {
    path: '/attendance/print',
    name: 'attendance-print',
    component: () => import('../views/AttendancePrint.vue'),
    meta: { title: '考勤统计' }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '工作台', icon: 'Odometer' }
      },
      {
        path: 'incoming',
        name: 'incoming',
        component: () => import('../views/IncomingDocs.vue'),
        meta: { title: '收文管理', icon: 'FolderOpened' }
      },
      {
        path: 'study',
        name: 'study',
        component: () => import('../views/Study.vue'),
        meta: { title: '公共资料', icon: 'Reading' }
      },
      {
        path: 'contacts',
        name: 'contacts',
        component: () => import('../views/Contacts.vue'),
        meta: { title: '通讯录', icon: 'Phone' }
      },
      {
        path: 'duty',
        name: 'duty',
        component: () => import('../views/Duty.vue'),
        meta: { title: '值守排班', icon: 'AlarmClock' }
      },
      {
        path: 'reports',
        name: 'reports',
        component: () => import('../views/Reports.vue'),
        meta: { title: '周/月/年报', icon: 'Tickets' }
      },
      {
        path: 'attendance',
        name: 'attendance',
        component: () => import('../views/Attendance.vue'),
        meta: { title: '考勤点到', icon: 'AlarmClock' }
      },
      {
        path: 'leave',
        name: 'leave',
        component: () => import('../views/Leave.vue'),
        meta: { title: '请假管理', icon: 'Calendar' }
      },
      {
        path: 'vehicles',
        name: 'vehicles',
        component: () => import('../views/Vehicles.vue'),
        meta: { title: '公车管理', icon: 'Van' }
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('../views/Users.vue'),
        meta: { title: '用户管理', icon: 'User', admin: true }
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('../views/Profile.vue'),
        meta: { title: '个人中心', icon: 'Setting' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    // 强制改密拦截：未修改默认密码前只能访问个人中心
    const mustChange = localStorage.getItem('must_change') === '1'
    if (mustChange && to.path !== '/profile') {
      next('/profile')
      return
    }
    if (to.meta.admin) {
      const user = JSON.parse(localStorage.getItem('user') || '{}')
      if (user.role_code !== 'admin') {
        next('/dashboard')
        return
      }
    }
    document.title = to.meta.title ? `${to.meta.title} - 伊宁县委宣传部` : '伊宁县委宣传部'
    next()
  }
})

export default router
