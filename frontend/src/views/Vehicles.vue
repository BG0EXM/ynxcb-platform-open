<template>
  <div>
    <el-alert type="info" :closable="false" class="rule-alert">
      <template #title>
        公车使用无需申请审批，出车前在此<b>报备</b>即可。报备后可一键打印《派车单》。
      </template>
    </el-alert>

    <el-row :gutter="16" class="mt-12">
      <el-col :span="6" v-for="c in statCards" :key="c.label">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <div class="stat-card">
            <div class="stat-num" :style="{ color: c.color }">{{ c.value }}</div>
            <div class="stat-label">{{ c.label }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-tabs v-model="activeTab" class="mt-12">
      <!-- 车辆管理（管理员） -->
      <el-tab-pane label="车辆管理" name="vehicles" v-if="authStore.isAdmin">
        <el-card shadow="never">
          <div class="toolbar">
            <div>
              <el-input v-model="vQuery.keyword" placeholder="搜索车牌/品牌/司机" clearable style="width:220px" @keyup.enter="loadVehicles" @clear="loadVehicles" />
              <el-select v-model="vQuery.status" placeholder="状态" clearable style="width:120px" class="ml-8" @change="loadVehicles">
                <el-option label="可用" :value="1" />
                <el-option label="维修中" :value="3" />
              </el-select>
            </div>
            <el-button type="primary" @click="openVehicleCreate">新增车辆</el-button>
          </div>
          <el-table :data="vehicles" stripe v-loading="loading" empty-text="暂无车辆">
            <el-table-column prop="plate_no" label="车牌号" width="110" />
            <el-table-column prop="brand" label="品牌型号" width="140" />
            <el-table-column prop="seats" label="座位" width="60" />
            <el-table-column prop="driver" label="司机" width="90" />
            <el-table-column prop="vin" label="车架号" width="150" show-overflow-tooltip />
            <el-table-column prop="insurance_date" label="保险到期" width="105" />
            <el-table-column label="状态" width="85">
              <template #default="{ row }">
                <el-tag size="small" :type="vStatusType(row.status)">{{ vStatusNames[row.status] }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openVehicleEdit(row)">编辑</el-button>
                <el-button link type="danger" size="small" @click="removeVehicle(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 用车报备 -->
      <el-tab-pane label="用车报备" name="apply">
        <el-card shadow="never">
          <div class="toolbar">
            <div>
              <el-date-picker v-model="queryDate" type="date" value-format="YYYY-MM-DD" placeholder="按用车日期筛选" style="width:160px" @change="loadApplies" />
              <el-button v-if="authStore.isAdmin" class="ml-8" :type="showAll ? 'primary' : 'info'" plain @click="toggleMine">
                {{ showAll ? '查看全部' : '只看我的' }}
              </el-button>
            </div>
            <div>
              <el-button type="primary" @click="openApplyCreate">用车报备</el-button>
              <el-button type="success" :icon="'Download'" class="ml-8" @click="exportApplies">导出Excel</el-button>
            </div>
          </div>
          <el-table :data="applies" stripe v-loading="loading" empty-text="暂无报备记录">
            <el-table-column prop="use_date" label="用车日期" width="110" />
            <el-table-column prop="vehicle_no" label="车牌号" width="110" />
            <el-table-column prop="user_name" label="用车人" width="90" />
            <el-table-column prop="purpose" label="事由" min-width="130" show-overflow-tooltip />
            <el-table-column prop="destination" label="目的地" width="120" show-overflow-tooltip />
            <el-table-column prop="use_time" label="用车时间" width="90" />
            <el-table-column prop="passengers" label="人数" width="60" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openPrint(row)">打印派车单</el-button>
                <el-button link type="danger" size="small" @click="removeApply(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 车辆编辑对话框 -->
    <el-dialog v-model="vDialogVisible" :title="vEditId ? '编辑车辆' : '新增车辆'" width="480px">
      <el-form :model="vForm" label-width="90px">
        <el-form-item label="车牌号" required>
          <el-input v-model="vForm.plate_no" placeholder="如：新F12345" />
        </el-form-item>
        <el-form-item label="品牌型号">
          <el-input v-model="vForm.brand" placeholder="如：大众帕萨特" />
        </el-form-item>
        <el-form-item label="座位数">
          <el-input-number v-model="vForm.seats" :min="1" :max="50" />
        </el-form-item>
        <el-form-item label="司机">
          <el-input v-model="vForm.driver" />
        </el-form-item>
        <el-form-item label="车架号(VIN)">
          <el-input v-model="vForm.vin" placeholder="车架号" />
        </el-form-item>
        <el-form-item label="发动机号">
          <el-input v-model="vForm.engine_no" placeholder="发动机号" />
        </el-form-item>
        <el-form-item label="保险到期">
          <el-date-picker v-model="vForm.insurance_date" type="date" value-format="YYYY-MM-DD" style="width:100%" placeholder="保险到期日期" />
        </el-form-item>
        <el-form-item label="年检到期">
          <el-date-picker v-model="vForm.inspect_date" type="date" value-format="YYYY-MM-DD" style="width:100%" placeholder="年检到期日期" />
        </el-form-item>
        <el-form-item label="登记日期">
          <el-date-picker v-model="vForm.register_date" type="date" value-format="YYYY-MM-DD" style="width:100%" placeholder="登记日期" />
        </el-form-item>
        <el-form-item label="购置日期">
          <el-date-picker v-model="vForm.purchase_at" type="date" value-format="YYYY-MM-DD" style="width:100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="vForm.status" style="width:100%">
            <el-option label="可用" :value="1" />
            <el-option label="维修中" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="vForm.note" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="vDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveVehicle">保存</el-button>
      </template>
    </el-dialog>

    <!-- 用车报备对话框 -->
    <el-dialog v-model="applyDialogVisible" title="用车报备" width="520px">
      <el-form :model="applyForm" label-width="90px">
        <el-form-item label="选择车辆" required>
          <el-select v-model="applyForm.vehicle_id" placeholder="选择车辆" style="width:100%">
            <el-option v-for="v in vehicles" :key="v.id"
              :label="v.plate_no + ' ' + (v.brand || '') + '（' + vStatusNames[v.status] + '）'" :value="v.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="用车人" required>
          <el-input v-model="applyForm.user_name" placeholder="用车人姓名" />
        </el-form-item>
        <el-form-item label="用车事由" required>
          <el-input v-model="applyForm.purpose" placeholder="填写用车事由" />
        </el-form-item>
        <el-form-item label="目的地">
          <el-input v-model="applyForm.destination" placeholder="如：伊犁州委宣传部" />
        </el-form-item>
        <el-form-item label="用车日期">
          <el-date-picker v-model="applyForm.use_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
        </el-form-item>
        <el-form-item label="用车时间">
          <el-input v-model="applyForm.use_time" placeholder="如：09:00-18:00" />
        </el-form-item>
        <el-form-item label="乘车人数">
          <el-input-number v-model="applyForm.passengers" :min="1" :max="50" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applyDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitApply">报备</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request, { exportFile } from '../utils/request'
import dayjs from 'dayjs'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
const activeTab = ref('vehicles')
const loading = ref(false)

// 车辆
const vehicles = ref([])
const vQuery = reactive({ keyword: '', status: '' })
const vDialogVisible = ref(false)
const vEditId = ref(0)
const vForm = reactive({ plate_no: '', brand: '', seats: 5, driver: '', status: 1, vin: '', engine_no: '', insurance_date: '', inspect_date: '', register_date: '', purchase_at: '', note: '' })

// 报备
const applies = ref([])
const queryDate = ref('')
const showAll = ref(false)
const applyDialogVisible = ref(false)
const applyForm = reactive({ vehicle_id: null, user_name: '', purpose: '', destination: '', use_date: dayjs().format('YYYY-MM-DD'), use_time: '', passengers: 1 })

const vStatusNames = { 1: '可用', 2: '使用中', 3: '维修中', 4: '已报废' }
const vStatusType = (s) => ({ 1: 'success', 2: 'warning', 3: 'danger', 4: 'info' }[s] || 'info')

const stats = reactive({ total: 0, available: 0, in_use: 0, repair: 0, today_use: 0 })

const statCards = computed(() => [
  { label: '车辆总数', value: stats.total, color: '#303133' },
  { label: '可用车辆', value: stats.available, color: '#67c23a' },
  { label: '维修中', value: stats.repair, color: '#f56c6c' },
  { label: '今日报备', value: stats.today_use, color: '#e6a23c' }
])

const loadVehicles = async () => {
  loading.value = true
  try {
    const res = await request.get('/vehicles', { params: vQuery })
    vehicles.value = res.list || []
  } catch (e) {
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const res = await request.get('/vehicle-stats')
    Object.assign(stats, res)
  } catch (e) {}
}

const loadApplies = async () => {
  loading.value = true
  try {
    const params = {}
    if (queryDate.value) params.date = queryDate.value
    if (!showAll.value) params.mine = '1'
    const res = await request.get('/vehicle-applies', { params })
    applies.value = res.list || []
  } catch (e) {
  } finally {
    loading.value = false
  }
}

const openVehicleCreate = () => {
  vEditId.value = 0
  Object.assign(vForm, { plate_no: '', brand: '', seats: 5, driver: '', status: 1, vin: '', engine_no: '', insurance_date: '', inspect_date: '', register_date: '', purchase_at: '', note: '' })
  vDialogVisible.value = true
}

const openVehicleEdit = (row) => {
  vEditId.value = row.id
  Object.assign(vForm, {
    plate_no: row.plate_no, brand: row.brand, seats: row.seats, driver: row.driver,
    status: row.status, vin: row.vin, engine_no: row.engine_no,
    insurance_date: row.insurance_date, inspect_date: row.inspect_date,
    register_date: row.register_date, purchase_at: row.purchase_at, note: row.note
  })
  vDialogVisible.value = true
}

const saveVehicle = async () => {
  if (!vForm.plate_no) return ElMessage.warning('请输入车牌号')
  try {
    if (vEditId.value) {
      await request.put('/vehicles', { ...vForm, id: vEditId.value })
      ElMessage.success('更新成功')
    } else {
      await request.post('/vehicles', vForm)
      ElMessage.success('添加成功')
    }
    vDialogVisible.value = false
    loadVehicles()
    loadStats()
  } catch (e) {}
}

const removeVehicle = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除车辆「${row.plate_no}」？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.delete(`/vehicles/${row.id}`)
    ElMessage.success('删除成功')
    loadVehicles()
    loadStats()
  } catch (e) {}
}

const openApplyCreate = () => {
  Object.assign(applyForm, {
    vehicle_id: null, user_name: '', purpose: '', destination: '',
    use_date: dayjs().format('YYYY-MM-DD'), use_time: '', passengers: 1
  })
  applyDialogVisible.value = true
}

const submitApply = async () => {
  if (!applyForm.vehicle_id) return ElMessage.warning('请选择车辆')
  if (!applyForm.user_name) return ElMessage.warning('请填写用车人')
  if (!applyForm.purpose) return ElMessage.warning('请填写用车事由')
  try {
    const res = await request.post('/vehicle-applies', applyForm)
    ElMessage.success(res.message || '报备成功')
    applyDialogVisible.value = false
    loadApplies()
    loadStats()
  } catch (e) {}
}

const removeApply = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除该条报备？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.delete(`/vehicle-applies/${row.id}`)
    ElMessage.success('删除成功')
    loadApplies()
  } catch (e) {}
}

const openPrint = (row) => {
  sessionStorage.setItem('printVehicle', JSON.stringify(row))
  const url = `/vehicle/print/${row.id}`
  window.open(url, '_blank')
}

const toggleMine = () => {
  showAll.value = !showAll.value
  loadApplies()
}

onMounted(() => {
  loadVehicles()
  loadApplies()
  loadStats()
})

const exportApplies = () => {
  const params = {}
  if (queryDate.value) params.date = queryDate.value
  exportFile('/export/vehicle-applies', params)
}
</script>

<style scoped>
.rule-alert {
  margin-bottom: 0;
}
.mt-12 {
  margin-top: 12px;
}
.stat-card {
  text-align: center;
}
.stat-num {
  font-size: 28px;
  font-weight: 700;
}
.stat-label {
  color: #909399;
  font-size: 13px;
  margin-top: 4px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}
.ml-8 {
  margin-left: 8px;
}
</style>
