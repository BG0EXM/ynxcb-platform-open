<template>
  <div>
    <div class="no-print toolbar">
      <el-button type="primary" :icon="'Printer'" @click="printPage">打印呈批单</el-button>
      <el-button type="success" :icon="'Tickets'" @click="printCard">打印传阅登记卡</el-button>
      <el-button :icon="'Back'" @click="goBack">返回</el-button>
    </div>

    <div class="a4-page" id="print-area">
      <div v-if="doc.need_return === 1" class="need-return-mark">
        {{ doc.returned === 1 ? '已退回' : '需退回' }}
      </div>
      <div class="doc-header">
        <span class="unit-name">中共伊宁县委宣传部</span>
      </div>
      <div class="doc-title">收文呈批单</div>
      <table class="approve-table">
        <tbody>
          <tr>
            <td class="label">来文单位</td>
            <td class="value">{{ doc.from_unit || '' }}</td>
            <td class="label">收文编号</td>
            <td class="value">{{ doc.receive_no || '' }}</td>
          </tr>
          <tr>
            <td class="label">来文字号</td>
            <td class="value">{{ doc.from_doc_no || '' }}</td>
            <td class="label">收文日期</td>
            <td class="value">{{ doc.received_date || '' }}</td>
          </tr>
          <tr>
            <td class="label">文件标题</td>
            <td class="value title-cell" colspan="3">{{ doc.title || '' }}</td>
          </tr>
          <tr>
            <td class="label">密级</td>
            <td class="value">{{ doc.secret_level || '' }}</td>
            <td class="label">紧急程度</td>
            <td class="value">{{ doc.urgency || '' }}</td>
          </tr>
          <tr>
            <td class="label">份数</td>
            <td class="value">{{ doc.copies }} 份</td>
            <td class="label">登记人</td>
            <td class="value">{{ doc.registrar_name || '' }}</td>
          </tr>
          <tr>
            <td class="label">拟办意见</td>
            <td class="value tall" colspan="3">
              <div class="cell-text">{{ doc.suggest || '' }}</div>
            </td>
          </tr>
          <tr>
            <td class="label">领导批示</td>
            <td class="value tall" colspan="3">
              <div class="cell-text">{{ doc.leader_comment || '' }}</div>
            </td>
          </tr>
          <tr>
            <td class="label">办理情况</td>
            <td class="value tall" colspan="3">
              <div class="cell-text">{{ doc.processing || '' }}</div>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="doc-footer">
        <span class="footer-left">打印日期：{{ today }}</span>
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
const doc = ref({})

const today = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
})

const goBack = () => router.push('/incoming')

const printPage = () => {
  window.print()
}

const printCard = () => {
  window.open(`/incoming/print-card/${route.params.id}`, '_blank')
}

onMounted(async () => {
  const cached = sessionStorage.getItem('printDoc')
  if (cached) {
    doc.value = JSON.parse(cached)
  } else {
    try {
      doc.value = await request.get(`/incoming-docs/${route.params.id}`)
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
.need-return-mark {
  position: absolute;
  top: 24px;
  right: 24px;
  border: 3px solid #c8102e;
  color: #c8102e;
  font-size: 22px;
  font-weight: 700;
  padding: 6px 16px;
  letter-spacing: 4px;
  transform: rotate(0deg);
  background: #fff;
}
.unit-name {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 2px;
}
.doc-title {
  text-align: center;
  font-size: 24px;
  font-weight: 700;
  margin: 18px 0 16px;
  letter-spacing: 6px;
}
.approve-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #000;
}
.approve-table td {
  border: 1px solid #000;
  padding: 10px 12px;
  font-size: 15px;
  line-height: 1.8;
}
.label {
  width: 110px;
  text-align: center;
  font-weight: 600;
  background: #f5f5f5;
  vertical-align: middle;
}
.value {
  width: 160px;
  vertical-align: middle;
}
.title-cell {
  font-weight: 600;
}
.tall {
  height: 110px;
}
.cell-text {
  white-space: pre-wrap;
  min-height: 90px;
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
