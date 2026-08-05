<template>
  <div>
    <el-alert type="info" :closable="false" class="rule-alert">
      <template #title>
        考勤点到：由管理员在每日晨会上登录平台，为当日到岗人员手工点到（出勤/请假/出差/未到）。
      </template>
    </el-alert>

    <div class="stats-entry">
      <el-button type="primary" :icon="'DataAnalysis'" @click="goStats">考勤统计（月度/年度，可打印）</el-button>
      <el-button type="success" :icon="'Download'" @click="exportData">导出Excel</el-button>
    </div>

    <!-- 管理员点到操作区 -->
    <el-card shadow="never" v-if="authStore.isAdmin" class="mt-12">
      <template #header>
        <div class="card-header">
          <span class="card-title">晨会点到</span>
          <div class="header-right">
            <el-date-picker v-model="markDate" type="date" value-format="YYYY-MM-DD" placeholder="选择点到日期" style="width:160px" @change="loadMarkUsers" />
            <el-button type="primary" :icon="'CircleCheck'" @click="markAllPresent" class="ml-8">全部出勤</el-button>
          </div>
        </div>
      </template>

      <div class="mark-area">
        <div v-if="!markUsers.length" class="mark-empty">
          <el-empty description="暂无人员" :image-size="60" />
        </div>
        <el-table v-else :data="markUsers" size="small">
          <el-table-column prop="real_name" label="姓名" width="100" />
          <el-table-column prop="department" label="部门" width="130" />
          <el-table-column label="状态" width="260">
            <template #default="{ row }">
              <el-radio-group v-model="row.status">
                <el-radio :label="1">出勤</el-radio>
                <el-radio :label="2">请假</el-radio>
                <el-radio :label="3">出差</el-radio>
                <el-radio :label="4">未到</el-radio>
              </el-radio-group>
              <el-select v-if="row.status === 2" v-model="row.leave_type" size="small" placeholder="请假类型" style="width:120px;margin-left:8px">
                <el-option v-for="(name, val) in leaveTypeNames" :key="val" :label="name" :value="val" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="备注" width="180">
            <template #default="{ row }">
              <el-input v-model="row.remark" size="small" placeholder="备注" />
            </template>
          </el-table-column>
        </el-table>
        <div class="mark-buttons" v-if="markUsers.length">
          <el-button type="primary" size="large" :loading="saving" @click="saveMark">保存点到结果</el-button>
        </div>
      </div>
    </el-card>

    <!-- 考勤记录与统计 -->
    <el-row :gutter="16" class="mt-12">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <span class="card-title">点到统计</span>
          </template>
          <div v-if="stats">
            <div class="stat-row"><span>日期</span><b>{{ stats.date }}</b></div>
            <div class="stat-row"><span>已点到</span><b>{{ stats.stats.total }} 人</b></div>
            <div class="stat-row"><span>出勤</span><b style="color:#67c23a">{{ stats.stats.present }} 人</b></div>
            <div class="stat-row"><span>请假</span><b style="color:#e6a23c">{{ stats.stats.leave }} 人</b></div>
            <div class="stat-row"><span>出差</span><b style="color:#409eff">{{ stats.stats.trip }} 人</b></div>
            <div class="stat-row"><span>未到</span><b style="color:#f56c6c">{{ stats.stats.absent }} 人</b></div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span class="card-title">考勤记录</span>
              <div class="header-right">
                <el-date-picker v-model="queryDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width:160px" @change="loadData" />
                <el-select v-if="authStore.isAdmin" v-model="userFilter" placeholder="全部人员" clearable style="width:140px" class="ml-8" @change="loadData">
                  <el-option v-for="a in assignees" :key="a.id" :label="a.real_name" :value="a.id" />
                </el-select>
              </div>
            </div>
          </template>
          <el-table :data="list" stripe v-loading="loading" empty-text="该日暂无考勤记录">
            <el-table-column prop="attend_date" label="日期" width="110" />
            <el-table-column v-if="authStore.isAdmin" prop="user_name" label="姓名" width="100" />
            <el-table-column label="状态" width="130">
              <template #default="{ row }">
                <el-tag size="small" :type="statusType(row.status)">{{ statusNames[row.status] }}</el-tag>
                <el-tag v-if="row.status === 2 && row.leave_type" size="small" type="info" class="ml-4">{{ leaveTypeNames[row.leave_type] }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import request, { exportFile } from '../utils/request'
import dayjs from 'dayjs'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const router = useRouter()
const markDate = ref(dayjs().format('YYYY-MM-DD'))
const queryDate = ref(dayjs().format('YYYY-MM-DD'))
const markUsers = ref([])
const list = ref([])
const stats = ref(null)
const loading = ref(false)
const saving = ref(false)
const userFilter = ref('')
const assignees = ref([])

const statusNames = { 1: '出勤', 2: '请假', 3: '出差', 4: '未到' }
const statusType = (s) => ({ 1: 'success', 2: 'warning', 3: 'primary', 4: 'danger' }[s] || 'info')
const leaveTypeNames = {
  annual: '年假', sick: '病假', personal: '事假', marriage: '婚假',
  maternity: '产假', bereavement: '丧假', other: '其他'
}

const loadMarkUsers = async () => {
  if (!authStore.isAdmin) return
  try {
    const res = await request.get('/attendance/mark-users', { params: { date: markDate.value } })
    markUsers.value = res.list || []
  } catch (e) {}
}

const markAllPresent = () => {
  markUsers.value.forEach(u => { u.status = 1 })
}

const saveMark = async () => {
  if (!markUsers.value.length) return ElMessage.warning('没有人员')
  saving.value = true
  try {
    const records = markUsers.value.map(u => ({
      user_id: u.id, status: u.status, leave_type: u.leave_type || '', remark: u.remark || ''
    }))
    await request.post('/attendance/mark', { attend_date: markDate.value, records })
    ElMessage.success('点到完成')
    loadMarkUsers()
    loadStats()
    loadData()
  } catch (e) {
  } finally {
    saving.value = false
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const params = { date: queryDate.value }
    if (userFilter.value) params.user_id = userFilter.value
    const res = await request.get('/attendance/list', { params })
    list.value = res.list || []
  } catch (e) {
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const res = await request.get('/attendance/stats', { params: { date: queryDate.value } })
    stats.value = res
  } catch (e) {}
}

const loadAssignees = async () => {
  try {
    const res = await request.get('/assignees')
    assignees.value = res.list || []
  } catch (e) {}
}

const goStats = () => {
  router.push('/attendance/print')
}

const exportData = () => {
  const params = {}
  if (queryDate.value) params.date = queryDate.value
  exportFile('/export/attendances', params)
}

onMounted(() => {
  if (authStore.isAdmin) {
    loadMarkUsers()
    loadAssignees()
  }
  loadData()
  loadStats()
})
</script>

<style scoped>
.rule-alert {
  margin-bottom: 0;
}
.stats-entry {
  margin-top: 12px;
}
.mt-12 {
  margin-top: 12px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-title {
  font-weight: 600;
}
.header-right {
  display: flex;
  align-items: center;
}
.ml-8 {
  margin-left: 8px;
}
.ml-4 {
  margin-left: 4px;
}
.mark-area {
  min-height: 100px;
}
.mark-empty {
  padding: 20px 0;
}
.mark-buttons {
  margin-top: 16px;
  text-align: right;
}
.stat-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px dashed #ebeef5;
  color: #606266;
}
.stat-row:last-child {
  border-bottom: none;
}
</style>
