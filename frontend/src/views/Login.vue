<template>
  <div class="login-page">
    <div class="login-box">
      <div class="login-header">
        <img src="../assets/danghui.png" alt="党徽" class="danghui-img" />
        <h1>伊宁县委宣传部</h1>
        <p>部务工作平台</p>
      </div>
      <el-form :model="form" :rules="rules" ref="formRef" size="large">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="'User'" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码"
            :prefix-icon="'Lock'" show-password @keyup.enter="handleLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-btn" :loading="loading" @click="handleLogin">
            登 录
          </el-button>
        </el-form-item>
      </el-form>
      <p class="login-tip">默认管理员账号：admin / admin123</p>
    </div>
    <div class="login-footer">
      <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener">ICP备案号占位</a>
      <span class="footer-sep">|</span>
      <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener">
        公网安备号占位
      </a>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../store/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref()
const loading = ref(false)
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  await formRef.value.validate()
  loading.value = true
  try {
    const res = await authStore.login(form.username, form.password)
    if (res.must_change) {
      ElMessage.warning('您正在使用默认密码，请先修改密码')
      router.push('/profile')
    } else {
      ElMessage.success('登录成功')
      router.push('/dashboard')
    }
  } catch (e) {
    // 错误已由拦截器提示
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #8f0f26;
  background-image: url('../assets/login-bg.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}
.login-box {
  width: 400px;
  padding: 48px 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
}
.login-header {
  text-align: center;
  margin-bottom: 32px;
}
.danghui-img {
  width: 64px;
  height: 64px;
  object-fit: contain;
  margin-bottom: 12px;
}
.login-header h1 {
  font-size: 24px;
  color: #c8102e;
  margin-bottom: 6px;
}
.login-header p {
  color: #909399;
  font-size: 14px;
  letter-spacing: 6px;
}
.login-btn {
  width: 100%;
  background: #c8102e;
  border-color: #c8102e;
}
.login-btn:hover {
  background: #a90d26;
  border-color: #a90d26;
}
.login-tip {
  text-align: center;
  color: #c0c4cc;
  font-size: 12px;
  margin-top: 12px;
}
.login-footer {
  position: fixed;
  bottom: 16px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.85);
}
.login-footer a {
  color: rgba(255, 255, 255, 0.85);
  text-decoration: none;
}
.login-footer a:hover {
  text-decoration: underline;
}
.footer-sep {
  margin: 0 8px;
  color: rgba(255, 255, 255, 0.6);
}
</style>
