<template>
  <div>
    <div class="no-print toolbar">
      <el-button type="primary" :icon="'Printer'" @click="printPage">打印派车单</el-button>
      <el-button :icon="'Back'" @click="goBack">返回</el-button>
    </div>

    <div class="a4-page" id="print-area">
      <div class="doc-header">
        <span class="unit-name">中共伊宁县委宣传部</span>
      </div>
      <div class="doc-title">派 车 单</div>

      <table class="dispatch-table">
        <tbody>
          <tr>
            <td class="label">车牌号</td>
            <td class="value">{{ detail.vehicle_no || '' }}</td>
            <td class="label">车辆型号</td>
            <td class="value">{{ detail.vehicle_brand || '' }}</td>
          </tr>
          <tr>
            <td class="label">司机</td>
            <td class="value">{{ detail.vehicle_driver || '' }}</td>
            <td class="label">乘车人数</td>
            <td class="value">{{ detail.passengers }} 人</td>
          </tr>
          <tr>
            <td class="label">用车人</td>
            <td class="value">{{ detail.user_name || '' }}</td>
            <td class="label">报备人</td>
            <td class="value">{{ detail.reporter || '' }}</td>
          </tr>
          <tr>
            <td class="label">用车日期</td>
            <td class="value">{{ detail.use_date || '' }}</td>
            <td class="label">用车时间</td>
            <td class="value">{{ detail.use_time || '' }}</td>
          </tr>
          <tr>
            <td class="label">目的地</td>
            <td class="value" colspan="3">{{ detail.destination || '' }}</td>
          </tr>
          <tr>
            <td class="label">用车事由</td>
            <td class="value tall" colspan="3">
              <div class="cell-text">{{ detail.purpose || '' }}</div>
            </td>
          </tr>
          <tr>
            <td class="label">领导签字</td>
            <td class="value tall" colspan="3"></td>
          </tr>
          <tr>
            <td class="label">司机签字</td>
            <td class="value tall" colspan="3"></td>
          </tr>
        </tbody>
      </table>

      <div class="note">
        <p>备注：1. 本派车单由用车人报备后生成，出车前交司机备查。 2. 用车结束后请及时销单。</p>
      </div>

      <div class="doc-footer">
        <span>打印日期：{{ today }}</span>
        <span class="footer-right">
          <div>中共伊宁县委宣传部办公室</div>
          <div>伊宁县委宣传部部务工作平台V1.2</div>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '../utils/request'

const route = useRoute()
const router = useRouter()
const detail = ref({})

const today = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
})

const goBack = () => router.push('/vehicles')

const printPage = () => {
  window.print()
}

onMounted(async () => {
  const cached = sessionStorage.getItem('printVehicle')
  if (cached) {
    detail.value = JSON.parse(cached)
  } else {
    try {
      detail.value = await request.get(`/vehicle-applies/${route.params.id}`)
    } catch (e) {}
  }
})
</script>

<style scoped>
.toolbar {
  max-width: 794px;
  margin: 0 auto 12px;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgba(0,0,0,.1);
  display: flex;
  gap: 8px;
}
.a4-page {
  width: 794px;
  min-height: 1122px;
  margin: 0 auto;
  padding: 36px 56px 76px;
  background: #fff;
  box-shadow: 0 2px 12px rgba(0,0,0,.12);
  position: relative;
  box-sizing: border-box;
}
.doc-header {
  text-align: center;
}
.unit-name {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 2px;
}
.doc-title {
  text-align: center;
  font-size: 26px;
  font-weight: 700;
  margin: 24px 0 24px;
  letter-spacing: 8px;
}
.dispatch-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #000;
}
.dispatch-table td {
  border: 1px solid #000;
  padding: 12px 14px;
  font-size: 15px;
  line-height: 1.8;
}
.label {
  width: 110px;
  text-align: center;
  font-weight: 600;
  background: #f5f5f5;
}
.value {
  width: 200px;
}
.tall {
  height: 120px;
}
.cell-text {
  white-space: pre-wrap;
  min-height: 100px;
}
.note {
  margin-top: 16px;
  font-size: 12px;
  color: #666;
  line-height: 1.8;
}
.doc-footer {
  position: absolute;
  bottom: 36px;
  left: 56px;
  right: 56px;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  font-size: 13px;
  color: #333;
}
.footer-right {
  text-align: center;
  line-height: 1.7;
}
@media print {
  @page {
    size: A4;
    margin: 0;
  }
  body {
    background: #fff !important;
  }
  .no-print {
    display: none !important;
  }
  .a4-page {
    width: 100%;
    min-height: 100vh;
    box-shadow: none;
    padding: 30px 44px 70px;
  }
}
</style>
