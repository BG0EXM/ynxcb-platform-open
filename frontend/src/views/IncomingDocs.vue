<template>
  <div>
    <el-card shadow="never">
      <div class="toolbar">
        <div>
          <el-input v-model="query.keyword" placeholder="搜索标题/来文单位/字号" clearable style="width: 240px"
            @keyup.enter="loadData" @clear="loadData">
            <template #append><el-button :icon="'Search'" @click="loadData" /></template>
          </el-input>
          <el-select v-model="query.status" placeholder="状态" clearable style="width: 120px" class="ml-8" @change="loadData">
            <el-option v-for="(name, val) in statusNames" :key="val" :label="name" :value="Number(val)" />
          </el-select>
          <el-select v-model="query.returned" placeholder="是否已退" clearable style="width: 110px" class="ml-8" @change="loadData">
            <el-option label="已退" :value="1" />
            <el-option label="未退" :value="0" />
          </el-select>
          <el-select v-model="query.need_return" placeholder="是否需退回" clearable style="width: 120px" class="ml-8" @change="loadData">
            <el-option label="需要退回" :value="1" />
            <el-option label="不需要退回" :value="0" />
          </el-select>
          <el-date-picker v-model="dateRange" type="daterange" value-format="YYYY-MM-DD"
            range-separator="至" start-placeholder="收文起始" end-placeholder="收文结束"
            class="ml-8" @change="loadData" />
        </div>
        <div>
          <el-button v-if="isOffice" type="primary" @click="openCreate">收文登记</el-button>
          <el-button type="success" :icon="'Download'" class="ml-8" @click="exportData">导出Excel</el-button>
          <el-button :icon="'Refresh'" circle class="ml-8" @click="loadData" />
        </div>
      </div>

      <el-table :data="list" stripe v-loading="loading" empty-text="暂无收文记录" @row-click="openDetail">
        <el-table-column prop="receive_no" label="收文编号" width="130" />
        <el-table-column prop="received_date" label="收文日期" width="110">
          <template #default="{ row }">{{ row.received_date }}</template>
        </el-table-column>
        <el-table-column prop="from_unit" label="来文单位" width="140" show-overflow-tooltip />
        <el-table-column prop="from_doc_no" label="来文字号" width="140" show-overflow-tooltip />
        <el-table-column prop="doc_no" label="文件编号" width="120" show-overflow-tooltip />
        <el-table-column prop="title" label="文件标题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="secret_level" label="密级" width="70">
          <template #default="{ row }">
            <el-tag size="small" :type="secretTag(row.secret_level)">{{ row.secret_level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="需退回" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.need_return === 1" size="small" type="warning">需要</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="已退" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.returned === 1" size="small" type="success">已退</el-tag>
            <el-tag v-else-if="row.need_return === 1 && row.returned === 0" size="small" type="danger">未退</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column prop="return_date" label="退回日期" width="110" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="statusType(row.status)">{{ statusNames[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right" @click.stop>
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
            <el-button link type="warning" size="small" @click="openPrint(row)">打印</el-button>
            <el-button link type="info" size="small" @click="openLabel(row)">标签</el-button>
            <el-button v-if="isOffice" link type="success" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="isOffice" link type="danger" size="small" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 收文登记/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="editId ? '编辑收文' : '收文登记'" width="640px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="收文编号">
          <el-input v-model="form.receive_no" placeholder="如：伊宣收〔2026〕12号" />
        </el-form-item>
        <el-form-item label="收文日期">
          <el-date-picker v-model="form.received_date" type="date" value-format="YYYY-MM-DD"
            style="width: 100%" placeholder="选择收文日期" />
        </el-form-item>
        <el-form-item label="来文单位" required>
          <el-input v-model="form.from_unit" placeholder="上级来文单位名称" />
        </el-form-item>
        <el-form-item label="来文字号">
          <el-input v-model="form.from_doc_no" placeholder="如：州宣发〔2026〕5号" />
        </el-form-item>
        <el-form-item label="文件编号">
          <el-input v-model="form.doc_no" placeholder="如：XCB-2026-001" />
        </el-form-item>
        <el-form-item label="文件标题" required>
          <el-input v-model="form.title" placeholder="文件标题" />
        </el-form-item>
        <el-form-item label="份数">
          <el-input-number v-model="form.copies" :min="1" :max="999" />
        </el-form-item>
        <el-form-item label="密级">
          <el-radio-group v-model="form.secret_level">
            <el-radio value="普通">普通</el-radio>
            <el-radio value="秘密">秘密</el-radio>
            <el-radio value="机密">机密</el-radio>
            <el-radio value="绝密">绝密</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="紧急程度">
          <el-radio-group v-model="form.urgency">
            <el-radio value="一般">一般</el-radio>
            <el-radio value="紧急">紧急</el-radio>
            <el-radio value="特急">特急</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="拟办意见">
          <el-input v-model="form.suggest" type="textarea" :rows="2" placeholder="拟办意见，如：拟请××同志阅示，拟转××科办理" />
        </el-form-item>
        <el-form-item label="领导批示">
          <el-input v-model="form.leader_comment" type="textarea" :rows="2" placeholder="领导批示内容" />
        </el-form-item>
        <el-form-item label="办理情况">
          <el-input v-model="form.processing" type="textarea" :rows="2" placeholder="办理情况记录" />
        </el-form-item>
        <el-form-item label="需退回">
          <el-radio-group v-model="form.need_return">
            <el-radio :label="1">需要退回</el-radio>
            <el-radio :label="0">不需要退回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="是否已退">
          <el-radio-group v-model="form.returned" :disabled="form.need_return !== 1">
            <el-radio :label="1">已退</el-radio>
            <el-radio :label="0">未退</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="退回日期" v-if="form.returned === 1">
          <el-date-picker v-model="form.return_date" type="date" value-format="YYYY-MM-DD"
            style="width: 100%" placeholder="选择退回日期" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 200px">
            <el-option v-for="(name, val) in statusNames" :key="val" :label="name" :value="Number(val)" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 详情对话框：含呈批单预览 + 传阅管理 -->
    <el-dialog v-model="detailVisible" :title="`收文详情 - ${detail.title}`" width="760px">
      <el-tabs v-model="detailTab">
        <el-tab-pane label="基础信息" name="info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="收文编号">{{ detail.receive_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="收文日期">{{ detail.received_date || '—' }}</el-descriptions-item>
            <el-descriptions-item label="来文单位">{{ detail.from_unit || '—' }}</el-descriptions-item>
            <el-descriptions-item label="来文字号">{{ detail.from_doc_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="文件编号">{{ detail.doc_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="密级">{{ detail.secret_level }}</el-descriptions-item>
            <el-descriptions-item label="紧急程度">{{ detail.urgency }}</el-descriptions-item>
            <el-descriptions-item label="份数">{{ detail.copies }}</el-descriptions-item>
            <el-descriptions-item label="需退回">{{ detail.need_return === 1 ? '需要' : '不需要' }}</el-descriptions-item>
            <el-descriptions-item label="是否已退">{{ detail.returned === 1 ? '已退' : '未退' }}</el-descriptions-item>
            <el-descriptions-item label="退回日期">{{ detail.return_date || '—' }}</el-descriptions-item>
            <el-descriptions-item label="登记人">{{ detail.registrar_name }}</el-descriptions-item>
            <el-descriptions-item label="拟办意见" :span="2">{{ detail.suggest || '—' }}</el-descriptions-item>
            <el-descriptions-item label="领导批示" :span="2">{{ detail.leader_comment || '—' }}</el-descriptions-item>
            <el-descriptions-item label="办理情况" :span="2">{{ detail.processing || '—' }}</el-descriptions-item>
          </el-descriptions>
          <div class="mt-16">
            <el-button type="primary" @click="openPrint(detail)">打印呈批单</el-button>
            <el-button type="warning" @click="openPrintCard(detail)">打印传阅登记卡</el-button>
            <el-button type="info" @click="openLabel(detail)">打印标签</el-button>
          </div>
        </el-tab-pane>
        <el-tab-pane v-if="isOffice" label="传阅登记" name="circ">
          <div class="circ-toolbar">
            <el-select v-model="circUser" filterable placeholder="选择传阅人" style="width: 260px">
              <el-option v-for="a in assignees" :key="a.id" :label="a.real_name + (a.department ? ' (' + a.department + ')' : '')" :value="a.id" />
            </el-select>
            <el-button type="primary" :icon="'Plus'" @click="addCirc">添加传阅人</el-button>
          </div>
          <el-table :data="detail.circulations || []" size="small">
            <el-table-column prop="order_no" label="序号" width="60" />
            <el-table-column prop="user_name" label="传阅人" width="120" />
            <el-table-column label="传阅日期" width="150">
              <template #default="{ row }">
                <el-date-picker v-model="row.read_date" type="date" value-format="YYYY-MM-DD" size="small"
                  placeholder="选择日期" @change="updateCirc(row)" />
              </template>
            </el-table-column>
            <el-table-column label="签字">
              <template #default="{ row }">
                <el-input v-model="row.signature" size="small" placeholder="签名" @change="updateCirc(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="70">
              <template #default="{ row }">
                <el-button link type="danger" size="small" @click="removeCirc(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request, { exportFile } from '../utils/request'
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()
// 是否办公室用户（办公室才能新增/编辑/删除收文）
const isOffice = computed(() => authStore.user?.department_name === '办公室')

const list = ref([])
const loading = ref(false)
const assignees = ref([])
const query = reactive({ keyword: '', status: '', returned: '', need_return: '' })
const dateRange = ref([])

const statusNames = { 1: '待登记', 2: '拟办中', 3: '待批示', 4: '办理中', 5: '已办结' }
const statusType = (s) => ({ 1: 'info', 2: 'primary', 3: 'warning', 4: 'primary', 5: 'success' }[s] || 'info')
const secretTag = (s) => ({ 普通: 'info', 秘密: 'warning', 机密: 'danger', 绝密: 'danger' }[s] || 'info')

const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailTab = ref('info')
const editId = ref(0)
const form = reactive({
  receive_no: '', received_date: '', from_unit: '', from_doc_no: '', doc_no: '', title: '',
  copies: 1, secret_level: '普通', urgency: '一般', suggest: '', leader_comment: '', processing: '',
  return_date: '', returned: 0, need_return: 0, status: 1
})
const detail = ref({})
const circUser = ref(null)

const loadData = async () => {
  loading.value = true
  try {
    const params = { ...query }
    if (dateRange.value && dateRange.value.length === 2) {
      params.start = dateRange.value[0]
      params.end = dateRange.value[1]
    }
    const res = await request.get('/incoming-docs', { params })
    list.value = res.list || []
  } catch (e) {
  } finally {
    loading.value = false
  }
}

const loadAssignees = async () => {
  try {
    const res = await request.get('/assignees')
    assignees.value = res.list || []
  } catch (e) {}
}

const openCreate = () => {
  editId.value = 0
  const today = new Date().toISOString().slice(0, 10)
  Object.assign(form, {
    receive_no: '', received_date: today, from_unit: '', from_doc_no: '', doc_no: '', title: '',
    copies: 1, secret_level: '普通', urgency: '一般', suggest: '', leader_comment: '', processing: '',
    return_date: '', returned: 0, need_return: 0, status: 1
  })
  dialogVisible.value = true
}

const openEdit = (row) => {
  editId.value = row.id
  Object.assign(form, {
    receive_no: row.receive_no, received_date: row.received_date, from_unit: row.from_unit,
    from_doc_no: row.from_doc_no, doc_no: row.doc_no, title: row.title, copies: row.copies,
    secret_level: row.secret_level, urgency: row.urgency, suggest: row.suggest,
    leader_comment: row.leader_comment, processing: row.processing,
    return_date: row.return_date, returned: row.returned, need_return: row.need_return,
    status: row.status
  })
  dialogVisible.value = true
}

const save = async () => {
  if (!form.title) return ElMessage.warning('请输入文件标题')
  if (!form.from_unit) return ElMessage.warning('请输入来文单位')
  try {
    if (editId.value) {
      await request.put('/incoming-docs', { ...form, id: editId.value })
      ElMessage.success('更新成功')
    } else {
      await request.post('/incoming-docs', form)
      ElMessage.success('登记成功')
    }
    dialogVisible.value = false
    loadData()
    emitIncomingChanged()
  } catch (e) {}
}

// 通知顶栏刷新待办数
const emitIncomingChanged = () => {
  window.dispatchEvent(new Event('incoming-changed'))
}

const openDetail = async (row) => {
  try {
    const res = await request.get(`/incoming-docs/${row.id}`)
    detail.value = res
    detailTab.value = 'info'
    detailVisible.value = true
  } catch (e) {}
}

const openPrint = (row) => {
  sessionStorage.setItem('printDoc', JSON.stringify(row))
  const url = `/incoming/print/${row.id}`
  window.open(url, '_blank')
}

const openPrintCard = (row) => {
  sessionStorage.setItem('printCard', JSON.stringify(row))
  const url = `/incoming/print-card/${row.id}`
  window.open(url, '_blank')
}

const openLabel = (row) => {
  sessionStorage.setItem('printLabel', JSON.stringify(row))
  const url = `/incoming/label/${row.id}`
  window.open(url, '_blank')
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除收文「${row.title}」？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.delete(`/incoming-docs/${row.id}`)
    ElMessage.success('删除成功')
    loadData()
    emitIncomingChanged()
  } catch (e) {}
}

const addCirc = async () => {
  if (!circUser.value) return ElMessage.warning('请选择传阅人')
  try {
    await request.post('/circulations', { doc_id: detail.value.id, user_id: circUser.value })
    circUser.value = null
    const res = await request.get(`/incoming-docs/${detail.value.id}`)
    detail.value = res
    ElMessage.success('已添加传阅人')
  } catch (e) {}
}

const updateCirc = async (row) => {
  try {
    await request.put('/circulations', { id: row.id, read_date: row.read_date, signature: row.signature })
  } catch (e) {}
}

const removeCirc = async (row) => {
  try {
    await ElMessageBox.confirm(`确认移除传阅人「${row.user_name}」？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.delete(`/circulations/${row.id}`)
    const res = await request.get(`/incoming-docs/${detail.value.id}`)
    detail.value = res
  } catch (e) {}
}

onMounted(() => {
  loadData()
  loadAssignees()
})

const exportData = () => {
  const params = {}
  if (query.keyword) params.keyword = query.keyword
  exportFile('/export/incoming-docs', params)
}
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}
.ml-8 { margin-left: 8px; }
.mt-16 { margin-top: 16px; }
.circ-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
</style>
