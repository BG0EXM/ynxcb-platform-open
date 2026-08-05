import { defineStore } from 'pinia'
import request from '../utils/request'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    mustChange: localStorage.getItem('must_change') === '1'
  }),
  getters: {
    isAdmin: (state) => state.user?.role_code === 'admin',
    isLoggedIn: (state) => !!state.token
  },
  actions: {
    async login(username, password) {
      const res = await request.post('/auth/login', { username, password })
      this.token = res.token
      this.user = res.user
      this.mustChange = !!res.must_change
      localStorage.setItem('token', res.token)
      localStorage.setItem('user', JSON.stringify(res.user))
      localStorage.setItem('must_change', res.must_change ? '1' : '0')
      return res
    },
    setMustChange(val) {
      this.mustChange = val
      localStorage.setItem('must_change', val ? '1' : '0')
    },
    logout() {
      this.token = ''
      this.user = null
      this.mustChange = false
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('must_change')
    }
  }
})
