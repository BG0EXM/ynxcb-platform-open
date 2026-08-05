<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16">
      <el-col :span="6" v-for="card in cards" :key="card.label">
        <el-card shadow="hover" class="stat-card" @click="card.path && $router.push(card.path)">
          <div class="stat-inner">
            <div class="stat-icon" :style="{ background: card.color }">
              <el-icon :size="28" color="#fff"><component :is="card.icon" /></el-icon>
            </div>
            <div>
              <div class="stat-num">{{ card.value }}</div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="mt-16">
      <!-- 左侧主区：最新收文 -->
      <el-col :span="16">
        <el-card shadow="never" header="最新收文">
          <el-table :data="latestIncoming" stripe empty-text="暂无收文记录">
            <el-table-column prop="received_date" label="收文日期" width="110" />
            <el-table-column prop="from_unit" label="来文单位" width="130" show-overflow-tooltip />
            <el-table-column prop="title" label="文件标题" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                <el-tag v-if="row.status < 5" size="small" type="warning" class="mr-8">待办</el-tag>
                {{ row.title }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button link type="primary" @click="$router.push('/incoming')">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 右侧 -->
      <el-col :span="8">
        <el-card shadow="never" header="今日值班">
          <div class="duty-info" v-if="stats.today_duty">
            <el-icon :size="40" color="#c8102e"><AlarmClock /></el-icon>
            <div>
              <div class="duty-name">
                {{ stats.today_duty }}
                <el-tag v-if="stats.today_duty_dawangyuan === 1" size="small" type="danger" effect="plain">县委大院</el-tag>
              </div>
              <div class="duty-date">{{ today }}</div>
              <div class="duty-tip">值守至21:00收文，无文件可提前回家</div>
            </div>
          </div>
          <el-empty v-else description="今日暂无值班安排" :image-size="70" />
        </el-card>

        <el-card shadow="never" header="我的考勤" class="mt-16">
          <div class="attendance-info">
            <div class="att-item">
              <span>今日状态</span>
              <el-tag :type="attStatusType">{{ attStatusText }}</el-tag>
            </div>
            <div class="att-item">
              <span>本月出勤</span>
              <b>{{ stats.month_present }} 天</b>
            </div>
            <div class="att-item">
              <span>本月请假</span>
              <b>{{ stats.month_leave }} 天</b>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 本周排班 -->
    <el-card shadow="never" class="mt-16">
      <template #header>
        <div class="card-header">
          <span class="card-title">本周排班</span>
          <el-button link type="primary" @click="$router.push('/duty')">去排班</el-button>
        </div>
      </template>
      <el-row :gutter="12" v-if="weekDuty.length">
        <el-col :span="3" v-for="d in weekDuty" :key="d.duty_date">
          <div class="duty-day" :class="{ today: d.duty_date === todayStr }">
            <div class="duty-day-date">{{ d.duty_date.slice(5) }}</div>
            <div class="duty-day-name">{{ d.user_name }}</div>
            <el-tag v-if="d.is_dawangyuan === 1" size="small" type="danger" effect="plain">大院</el-tag>
          </div>
        </el-col>
      </el-row>
      <el-empty v-else description="本周暂无排班" :image-size="50" />
    </el-card>

    <!-- 快捷操作 -->
    <el-card shadow="never" class="mt-16">
      <template #header>
        <span class="card-title">快捷操作</span>
      </template>
      <div class="quick-actions">
        <el-button v-for="act in quickActions" :key="act.label" @click="act.path && $router.push(act.path)">
          {{ act.label }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import request from '../utils/request'
import dayjs from 'dayjs'

const stats = ref({})
const latestIncoming = ref([])
const weekDuty = ref([])
const today = new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })
const todayStr = dayjs().format('YYYY-MM-DD')

const cards = computed(() => [
  { label: '本月出勤', value: stats.value.month_present || 0, icon: 'Calendar', color: '#67c23a', path: '/attendance' },
  { label: '本月请假', value: stats.value.month_leave_days || 0, icon: 'Tickets', color: '#e6a23c', path: '/leave' },
  { label: '今日用车', value: stats.value.today_vehicle || 0, icon: 'Van', color: '#409eff', path: '/vehicles' },
  { label: '待办收文', value: stats.value.pending_incoming || 0, icon: 'FolderOpened', color: '#f56c6c', path: '/incoming' }
])

const attStatusText = computed(() => {
  const s = stats.value.today_attendance
  if (s === 1) return '已出勤'
  if (s === 2) return '已请假'
  if (s === 3) return '出差中'
  if (s === 4) return '未到'
  return '未点到'
})
const attStatusType = computed(() => {
  const s = stats.value.today_attendance
  if (s === 1) return 'success'
  if (s === 2) return 'warning'
  if (s === 3) return 'primary'
  if (s === 4) return 'danger'
  return 'info'
})

const quickActions = [
  { label: '收文登记', path: '/incoming' },
  { label: '用车报备', path: '/vehicles' },
  { label: '提交周报', path: '/reports' },
  { label: '考勤点到', path: '/attendance' }
]

onMounted(async () => {
  try {
    const res = await request.get('/dashboard-stats')
    stats.value = res
    latestIncoming.value = res.latest_incoming || []
    weekDuty.value = res.week_duty || []
  } catch (e) {}
})
</script>

<style scoped>
.stat-card {
  cursor: pointer;
}
.stat-inner {
  display: flex;
  align-items: center;
  gap: 16px;
}
.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-num {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
}
.stat-label {
  font-size: 13px;
  color: #909399;
}
.mt-16 {
  margin-top: 16px;
}
.mr-8 {
  margin-right: 8px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  font-weight: 600;
}
.duty-info {
  display: flex;
  align-items: center;
  gap: 16px;
}
.duty-name {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}
.duty-date {
  color: #909399;
  font-size: 13px;
  margin-top: 4px;
}
.duty-tip {
  color: #b88230;
  font-size: 12px;
  margin-top: 4px;
}
.attendance-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.att-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #606266;
  font-size: 14px;
}
.duty-day {
  text-align: center;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 10px 4px;
}
.duty-day.today {
  background: #fdf0f2;
  border-color: #c8102e;
}
.duty-day-date {
  font-size: 12px;
  color: #909399;
}
.duty-day-name {
  font-weight: 600;
  color: #303133;
  margin: 6px 0;
}
.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
