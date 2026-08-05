<template>
  <div>
    <div class="no-print toolbar">
      <el-button type="primary" :icon="'Printer'" @click="printPage">打印传阅登记卡</el-button>
      <el-button :icon="'Back'" @click="goBack">返回</el-button>
    </div>

    <div class="a4-page" id="print-area">
      <div v-if="doc.need_return === 1" class="need-return-mark">
        {{ doc.returned === 1 ? '已退回' : '需退回' }}
      </div>
      <div class="doc-header">
        <span class="unit-name">中共伊宁县委宣传部</span>
      </div>
      <div class="doc-title">文件传阅登记卡</div>

      <table class="info-table">
        <tbody>
          <tr>
            <td class="label">收文编号</td>
            <td class="value">{{ doc.receive_no || '' }}</td>
            <td class="label">来文单位</td>
            <td class="value">{{ doc.from_unit || '' }}</td>
          </tr>
          <tr>
            <td class="label">来文字号</td>
            <td class="value">{{ doc.from_doc_no || '' }}</td>
            <td class="label">收文日期</td>
            <td class="value">{{ doc.received_date || '' }}</td>
          </tr>
          <tr>
            <td class="label">文件标题</td>
            <td class="value" colspan="3" style="font-weight:600">{{ doc.title || '' }}</td>
          </tr>
        </tbody>
      </table>

      <!-- 传阅登记表：序号/签字手写，签退列 -->
      <div class="circ-wrap">
        <table class="circ-table">
          <thead>
            <tr>
              <th style="width:70px">序号</th>
              <th style="width:160px">传阅人</th>
              <th>签字</th>
              <th style="width:150px">传阅日期</th>
              <th style="width:90px">签退</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(c, i) in rows" :key="i" class="blank-row">
              <td class="center"></td>
              <td class="center">{{ c.user_name || '' }}</td>
              <td class="center"></td>
              <td class="center"></td>
              <td class="center"></td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="note">
        <p>备注：本文件共 {{ doc.copies || 1 }} 份，传阅完毕后请及时签退归档。</p>
      </div>

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
const circs = ref([])

const today = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
})

// 至少 16 行空白行，让表格填满 A4 纸；已有传阅人则填充姓名，其余留空手写
const rows = computed(() => {
  const list = []
  const count = Math.max(16, circs.value.length)
  for (let i = 0; i < count; i++) {
    list.push({ user_name: circs.value[i] ? circs.value[i].user_name : '' })
  }
  return list
})

const goBack = () => router.push('/incoming')

const printPage = () => {
  window.print()
}

onMounted(async () => {
  try {
    const res = await request.get(`/incoming-docs/${route.params.id}`)
    doc.value = res
    circs.value = res.circulations || []
  } catch (e) {}
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
  min-height: 1122px;   /* A4 297mm */
  margin: 0 auto;
  padding: 40px 56px 60px;
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
  margin: 20px 0 20px;
  letter-spacing: 6px;
}
.info-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #000;
  margin-bottom: 12px;
}
.info-table td {
  border: 1px solid #000;
  padding: 7px 12px;
  font-size: 14px;
  line-height: 1.6;
}
.info-table .label {
  width: 110px;
  text-align: center;
  font-weight: 600;
  background: #f5f5f5;
}
.circ-wrap {
  /* 占满头部以下的剩余空间，让表格延伸到底部 */
  min-height: 780px;
}
.circ-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #000;
  table-layout: fixed;
}
.circ-table th,
.circ-table td {
  border: 1px solid #000;
  padding: 8px;
  font-size: 14px;
}
.circ-table th {
  background: #f5f5f5;
  font-weight: 600;
}
.circ-table tbody tr {
  height: 44px;
}
.blank-row td {
  height: 44px;
}
.center {
  text-align: center;
}
.note {
  margin-top: 14px;
  font-size: 13px;
  color: #333;
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
    size: A4 portrait;
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
    padding: 36px 44px 50px;
  }
  .circ-wrap {
    min-height: auto;
  }
  /* 打印时确保表格延伸到页面底部 */
  .circ-table {
    height: calc(100vh - 280px);
  }
  .circ-table tbody tr {
    height: auto;
  }
  .blank-row td {
    height: 44px;
  }
}
</style>
