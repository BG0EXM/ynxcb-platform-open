<template>
  <div>
    <el-card shadow="never">
      <div class="toolbar">
        <div>
          <el-input v-model="keyword" placeholder="搜索姓名/用户名" clearable style="width: 220px"
            @keyup.enter="loadData" @clear="loadData">
            <template #append><el-button :icon="'Search'" @click="loadData" /></template>
          </el-input>
        </div>
        <el-button type="primary" @click="openCreate">新建用户</el-button>
      </div>

      <el-table :data="list" stripe v-loading="loading" empty-text="暂无用户">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="130" />
        <el-table-column prop="real_name" label="姓名" width="120" />
        <el-table-column prop="department_name" label="部门" width="140" />
        <el-table-column prop="role_name" label="角色" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="roleTag(row.role_name)">{{ row.role_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="电话" width="140" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link type="warning" size="small" @click="resetPwd(row)">重置密码</el-button>
            <el-button link type="danger" size="small" @click="toggleStatus(row)">
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button link type="danger" size="small" @click="removeUser(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editId ? '编辑用户' : '新建用户'" width="480px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" :disabled="!!editId" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="form.real_name" />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="form.department_id" style="width: 100%">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role_id" style="width: 100%">
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
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
import { useAuthStore } from '../store/auth'

const authStore = useAuthStore()

const list = ref([])
const loading = ref(false)
const keyword = ref('')
const departments = ref([])
const roles = ref([])
const dialogVisible = ref(false)
const editId = ref(0)
const form = reactive({ username: '', real_name: '', department_id: null, role_id: 2, phone: '' })

const roleTag = (r) => ({ '系统管理员': 'danger', '科室工作人员': 'primary', '乡镇/通讯员': 'success' }[r] || 'info')

const loadData = async () => {
  loading.value = true
  try {
    const res = await request.get('/users', { params: { keyword: keyword.value } })
    list.value = res.list || []
  } catch (e) {
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  try {
    const [d, r] = await Promise.all([request.get('/departments'), request.get('/roles')])
    departments.value = d.list || []
    roles.value = r.list || []
  } catch (e) {}
}

const openCreate = () => {
  editId.value = 0
  Object.assign(form, { username: '', real_name: '', department_id: null, role_id: 2, phone: '' })
  dialogVisible.value = true
}

const openEdit = (row) => {
  editId.value = row.id
  Object.assign(form, {
    username: row.username, real_name: row.real_name,
    department_id: row.department_id, role_id: row.role_id, phone: row.phone
  })
  dialogVisible.value = true
}

const save = async () => {
  if (!form.username || !form.real_name) return ElMessage.warning('用户名和姓名必填')
  try {
    if (editId.value) {
      await request.put('/users', { ...form, id: editId.value, status: 1 })
      ElMessage.success('更新成功')
    } else {
      const res = await request.post('/users', form)
      ElMessage.success(res.message || '创建成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (e) {}
}

const resetPwd = async (row) => {
  try {
    await ElMessageBox.confirm(`确认重置「${row.real_name}」的密码为 123456？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.post('/users/reset-password', { id: row.id })
    ElMessage.success('密码已重置')
  } catch (e) {}
}

const toggleStatus = async (row) => {
  const newStatus = row.status === 1 ? 0 : 1
  try {
    await ElMessageBox.confirm(`确认${newStatus === 1 ? '启用' : '禁用'}「${row.real_name}」？`, '提示', { type: 'warning' })
  } catch (e) { return }
  try {
    await request.put('/users', { ...row, status: newStatus })
    ElMessage.success('操作成功')
    loadData()
  } catch (e) {}
}

const removeUser = async (row) => {
  if (row.username === authStore.user?.username) {
    return ElMessage.warning('不能删除当前登录账号')
  }
  try {
    await ElMessageBox.confirm(
      `确认永久删除用户「${row.real_name}」（${row.username}）？\n删除后无法恢复，其考勤/请假/排班/报备等关联记录将一并清除。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
  } catch (e) { return }
  try {
    await request.delete(`/users/${row.id}`)
    ElMessage.success('用户已删除')
    loadData()
  } catch (e) {}
}

onMounted(() => {
  loadData()
  loadOptions()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}
</style>
