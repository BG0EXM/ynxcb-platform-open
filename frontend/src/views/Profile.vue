<template>
  <div class="profile">
    <el-alert v-if="authStore.mustChange" type="warning" :closable="false" class="must-change-alert"
      title="您正在使用默认密码，为保障账号安全，请立即修改密码后再使用系统功能。" />
    <el-card shadow="never" class="profile-card">
      <div class="profile-header">
        <el-avatar :size="64" class="avatar">{{ avatarText }}</el-avatar>
        <div>
          <div class="name">{{ authStore.user?.real_name }}</div>
          <div class="desc">{{ authStore.user?.role_name }} · {{ authStore.user?.department_name }}</div>
        </div>
      </div>
      <el-descriptions :column="2" border class="mt-16">
        <el-descriptions-item label="用户名">{{ authStore.user?.username }}</el-descriptions-item>
        <el-descriptions-item label="电话">{{ authStore.user?.phone || '未填写' }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ authStore.user?.department_name }}</el-descriptions-item>
        <el-descriptions-item label="角色">{{ authStore.user?.role_name }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">修改密码</el-divider>
      <el-form :model="pwdForm" label-width="100px" class="pwd-form">
        <el-form-item label="原密码">
          <el-input v-model="pwdForm.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="pwdForm.confirm" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="changePwd">确认修改</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../store/auth'
import request from '../utils/request'

const authStore = useAuthStore()
const pwdForm = reactive({ old_password: '', new_password: '', confirm: '' })

const avatarText = computed(() => {
  const name = authStore.user?.real_name || '用户'
  return name.charAt(name.length - 1)
})

const changePwd = async () => {
  if (!pwdForm.old_password || !pwdForm.new_password) return ElMessage.warning('请填写完整')
  if (pwdForm.new_password.length < 6) return ElMessage.warning('新密码至少6位')
  if (pwdForm.new_password !== pwdForm.confirm) return ElMessage.warning('两次密码不一致')
  try {
    await request.post('/auth/change-password', {
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password
    })
    ElMessage.success('密码修改成功')
    // 清除强制改密标记
    if (authStore.mustChange) {
      authStore.setMustChange(false)
      ElMessage.success('密码已更新，现在可以使用全部功能')
    }
    Object.assign(pwdForm, { old_password: '', new_password: '', confirm: '' })
  } catch (e) {}
}
</script>

<style scoped>
.profile {
  max-width: 800px;
  margin: 0 auto;
}
.must-change-alert {
  margin-bottom: 12px;
}
.profile-header {
  display: flex;
  align-items: center;
  gap: 20px;
}
.avatar {
  background: #c8102e;
  font-size: 24px;
}
.name {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}
.desc {
  color: #909399;
  margin-top: 4px;
}
.mt-16 {
  margin-top: 20px;
}
.pwd-form {
  max-width: 420px;
}
</style>
