<template>
  <div>
    <el-card shadow="never">
      <div class="toolbar">
        <div>
          <el-input v-model="keyword" placeholder="搜索姓名/职务/电话" clearable style="width: 240px"
            @keyup.enter="loadData" @clear="loadData">
            <template #append><el-button :icon="'Search'" @click="loadData" /></template>
          </el-input>
          <el-select v-model="department_id" placeholder="按部门" clearable style="width: 160px" class="ml-8" @change="loadData">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </div>
        <el-button type="primary" @click="openCreate">添加联系人</el-button>
      </div>

      <el-table :data="list" stripe v-loading="loading" empty-text="暂无联系人">
        <el-table-column prop="name" label="姓名" width="110" />
        <el-table-column prop="position" label="职务" width="150" />
        <el-table-column prop="department_name" label="部门" width="140" />
        <el-table-column prop="phone" label="电话" width="150" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" size="small" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editId ? '编辑联系人' : '添加联系人'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="姓名" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="职务">
          <el-input v-model="form.position" />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="form.department_id" style="width: 100%">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="form.phone" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../utils/request'

const list = ref([])
const loading = ref(false)
const keyword = ref('')
const department_id = ref('')
const departments = ref([])
const dialogVisible = ref(false)
const editId = ref(0)
const form = reactive({ name: '', position: '', department_id: null, phone: '' })

const loadData = async () => {
  loading.value = true
  try {
    const res = await request.get('/contacts', { params: { keyword: keyword.value, department_id: department_id.value } })
    list.value = res.list || []
  } catch (e) {
  } finally {
    loading.value = false
  }
}

const loadDepartments = async () => {
  try {
    const res = await request.get('/departments')
    departments.value = res.list || []
  } catch (e) {}
}

const openCreate = () => {
  editId.value = 0
  Object.assign(form, { name: '', position: '', department_id: null, phone: '' })
  dialogVisible.value = true
}

const openEdit = (row) => {
  editId.value = row.id
  Object.assign(form, { name: row.name, position: row.position, department_id: row.department_id, phone: row.phone })
  dialogVisible.value = true
}

const save = async () => {
  if (!form.name) return ElMessage.warning('请输入姓名')
  try {
    if (editId.value) {
      await request.put('/contacts', { ...form, id: editId.value })
      ElMessage.success('更新成功')
    } else {
      await request.post('/contacts', form)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (e) {}
}

const remove = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除联系人「${row.name}」？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.delete(`/contacts/${row.id}`)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {}
}

onMounted(() => {
  loadData()
  loadDepartments()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}
.ml-8 { margin-left: 8px; }
</style>
