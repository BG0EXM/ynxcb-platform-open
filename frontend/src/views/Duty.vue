<template>
  <div>
    <!-- 值守规则说明 -->
    <el-alert type="info" :closable="false" class="rule-alert">
      <template #title>
        <div class="rule-content">
          <b>值守规则说明：</b>当天值守至晚上21:00（收文结束）。若当日无文件接收，可提前回家。
          一天可安排<b>一至两人</b>值守。若当日同时有<b>县委大院排班</b>，请在排班时勾选对应选项。
        </div>
      </template>
    </el-alert>

    <el-card shadow="never" class="mt-12">
      <div class="toolbar">
        <div>
          <el-date-picker v-model="currentMonth" type="month" value-format="YYYY-MM"
            placeholder="选择月份" @change="loadData" />
          <span class="ml-8 tip">{{ authStore.isAdmin ? '点击日历格子安排当日值守人员' : '排班仅供查看，由管理员安排' }}</span>
        </div>
        <div>
          <el-tag type="warning" effect="light" class="mr-8">值守至21:00收文</el-tag>
          <el-button type="success" :icon="'Download'" @click="exportData">导出Excel</el-button>
        </div>
      </div>

      <div class="duty-calendar">
        <div class="week-header">
          <div v-for="w in weekNames" :key="w" class="week-cell">{{ w }}</div>
        </div>
        <div class="duty-grid">
          <div v-for="cell in calendarCells" :key="cell.dateStr" class="day-cell"
            :class="{ 'empty': !cell.date, 'today': cell.isToday }">
            <template v-if="cell.date">
              <div class="day-num">{{ cell.day }}</div>
              <div class="duty-box" :class="[cell.schedules.length ? 'assigned' : 'empty-shift', { 'clickable': authStore.isAdmin }]"
                @click="authStore.isAdmin && openEdit(cell.dateStr)">
                <template v-if="cell.schedules.length">
                  <div v-for="s in cell.schedules" :key="s.id" class="duty-person">
                    <span class="duty-name">{{ s.user_name }}</span>
                    <el-tag v-if="s.is_dawangyuan === 1" size="small" type="danger" effect="plain">大院</el-tag>
                  </div>
                </template>
                <template v-else>
                  <div class="duty-empty">{{ authStore.isAdmin ? '安排值守' : '未排班' }}</div>
                </template>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- 图例 -->
      <div class="legend">
        <span class="legend-item"><el-tag size="small" type="danger" effect="plain">县委大院</el-tag> 当天同时有县委大院排班</span>
      </div>
    </el-card>

    <!-- 排班编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="`安排值守 - ${editDate}`" width="520px">
      <div class="dialog-rule">
        当天值守至 21:00（收文结束），无文件可提前回家。一天最多安排两人值守。
      </div>

      <!-- 当天已排班人员 -->
      <div class="sched-section">
        <div class="sched-title">已排班人员（{{ daySchedules.length }}人）</div>
        <el-table :data="daySchedules" size="small" empty-text="当日暂无排班">
          <el-table-column prop="user_name" label="值守人员" width="120" />
          <el-table-column label="县委大院" width="110">
            <template #default="{ row }">
              <el-checkbox :model-value="row.is_dawangyuan === 1" @change="(v) => toggleDaWangYuan(row, v)" />
            </template>
          </el-table-column>
          <el-table-column prop="note" label="备注" show-overflow-tooltip />
          <el-table-column label="操作" width="70">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="removeOne(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 添加人员 -->
      <div class="add-section">
        <div class="sched-title">添加值守人员</div>
        <div class="add-row">
          <el-select v-model="addUser" filterable placeholder="选择值守人员" style="flex: 1">
            <el-option v-for="a in availableAssignees" :key="a.id"
              :label="a.real_name + (a.department ? ' (' + a.department + ')' : '')" :value="a.id" />
          </el-select>
          <el-button type="primary" :icon="'Plus'" @click="addSchedule" :disabled="daySchedules.length >= 2">添加</el-button>
        </div>
        <div v-if="daySchedules.length >= 2" class="max-tip">一天最多安排两人值守</div>
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request, { exportFile } from '../utils/request'
import dayjs from 'dayjs'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const currentMonth = ref(dayjs().format('YYYY-MM'))
const schedules = ref([])
const assignees = ref([])
const dialogVisible = ref(false)
const editDate = ref('')
const daySchedules = ref([])
const addUser = ref(null)

const weekNames = ['日', '一', '二', '三', '四', '五', '六']

// 可添加的人员（未在该日排班的）
const availableAssignees = computed(() => {
  const assignedIds = new Set(daySchedules.value.map(s => s.user_id))
  return assignees.value.filter(a => !assignedIds.has(a.id))
})

const calendarCells = computed(() => {
  const [year, month] = currentMonth.value.split('-').map(Number)
  const firstDay = dayjs(`${year}-${month}-01`)
  const startOffset = firstDay.day()
  const daysInMonth = firstDay.daysInMonth()
  const cells = []
  for (let i = 0; i < startOffset; i++) {
    cells.push({ date: null })
  }
  const scheduleMap = {}
  schedules.value.forEach(s => {
    if (!scheduleMap[s.duty_date]) scheduleMap[s.duty_date] = []
    scheduleMap[s.duty_date].push(s)
  })
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${year}-${String(month).padStart(2, '0')}-${String(d).padStart(2, '0')}`
    cells.push({
      date: dateStr,
      day: d,
      isToday: dayjs().format('YYYY-MM-DD') === dateStr,
      dateStr,
      schedules: scheduleMap[dateStr] || []
    })
  }
  return cells
})

const loadData = async () => {
  try {
    const res = await request.get('/duty-schedules', { params: { month: currentMonth.value } })
    schedules.value = res.list || []
  } catch (e) {}
}

const loadAssignees = async () => {
  try {
    const res = await request.get('/assignees')
    assignees.value = res.list || []
  } catch (e) {}
}

const openEdit = (date) => {
  editDate.value = date
  daySchedules.value = schedules.value.filter(s => s.duty_date === date)
  addUser.value = null
  dialogVisible.value = true
}

const addSchedule = async () => {
  if (!addUser.value) return ElMessage.warning('请选择值守人员')
  if (daySchedules.value.length >= 2) return ElMessage.warning('一天最多安排两人值守')
  try {
    await request.post('/duty-schedules', {
      duty_date: editDate.value,
      user_id: addUser.value,
      is_dawangyuan: 0,
      note: ''
    })
    ElMessage.success('已添加')
    addUser.value = null
    await loadData()
    daySchedules.value = schedules.value.filter(s => s.duty_date === editDate.value)
  } catch (e) {}
}

const toggleDaWangYuan = async (row, val) => {
  try {
    await request.post('/duty-schedules', {
      duty_date: row.duty_date,
      user_id: row.user_id,
      is_dawangyuan: val ? 1 : 0,
      note: row.note || ''
    })
    ElMessage.success(val ? '已标记县委大院排班' : '已取消县委大院标记')
    await loadData()
    daySchedules.value = schedules.value.filter(s => s.duty_date === editDate.value)
  } catch (e) {}
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm(`确认移除 ${editDate.value} 的「${row.user_name}」排班？`, '删除确认', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.delete(`/duty-schedules/${row.id}`)
    ElMessage.success('已删除')
    await loadData()
    daySchedules.value = schedules.value.filter(s => s.duty_date === editDate.value)
  } catch (e) {}
}

onMounted(() => {
  loadData()
  loadAssignees()
})

const exportData = () => {
  const params = {}
  if (currentMonth.value) params.month = currentMonth.value
  exportFile('/export/duty-schedules', params)
}
</script>

<style scoped>
.rule-alert {
  margin-bottom: 0;
}
.rule-content {
  line-height: 1.8;
  color: #606266;
}
.mt-12 {
  margin-top: 12px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.ml-8 { margin-left: 8px; }
.tip { color: #909399; font-size: 13px; }
.duty-calendar {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}
.week-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}
.week-cell {
  padding: 10px;
  text-align: center;
  font-weight: 600;
  color: #606266;
  border-right: 1px solid #e4e7ed;
}
.week-cell:last-child { border-right: none; }
.duty-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
}
.day-cell {
  min-height: 100px;
  border-right: 1px solid #e4e7ed;
  border-bottom: 1px solid #e4e7ed;
  padding: 6px;
  position: relative;
}
.day-cell:nth-child(7n) { border-right: none; }
.day-cell.empty { background: #fafafa; }
.day-cell.today {
  background: #fdf0f2;
}
.day-num {
  font-size: 14px;
  color: #303133;
  margin-bottom: 6px;
  font-weight: 600;
}
.today .day-num {
  color: #c8102e;
}
.duty-box {
  border-radius: 4px;
  padding: 6px;
  min-height: 46px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.duty-box.clickable {
  cursor: pointer;
}
.duty-box.assigned {
  background: #e8f4fd;
  border: 1px solid #a0cfff;
}
.duty-box.empty-shift {
  background: #f5f7fa;
  border: 1px dashed #dcdfe6;
  align-items: center;
  justify-content: center;
}
.duty-name {
  font-weight: 600;
  color: #303133;
  font-size: 13px;
}
.duty-person {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}
.duty-empty {
  color: #c0c4cc;
  font-size: 12px;
}
.legend {
  margin-top: 12px;
  display: flex;
  gap: 24px;
  font-size: 13px;
  color: #606266;
}
.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.dialog-rule {
  background: #fdf6ec;
  border: 1px solid #faecd8;
  color: #b88230;
  border-radius: 4px;
  padding: 8px 12px;
  font-size: 13px;
  margin-bottom: 16px;
}
.sched-section {
  margin-bottom: 16px;
}
.sched-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}
.add-section {
  border-top: 1px solid #ebeef5;
  padding-top: 12px;
}
.add-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.max-tip {
  margin-top: 6px;
  color: #e6a23c;
  font-size: 12px;
}
</style>
