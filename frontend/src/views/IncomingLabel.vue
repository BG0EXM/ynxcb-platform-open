<template>
  <div>
    <div class="no-print toolbar">
      <el-button type="primary" :icon="'Printer'" @click="printPage">打印标签</el-button>
      <el-button :icon="'Back'" @click="goBack">返回</el-button>
    </div>
    <div class="no-print print-tips">
      <p>打印设置（HPRT D35BT）：</p>
      <p>1. 点「打印标签」，在系统打印对话框中「目标打印机」选 D35BT</p>
      <p>2. 「方向」选「横向」</p>
      <p>3. 「纸张大小」选自定义「80×50」（宽度80mm、长度50mm）</p>
      <p>4. 「旋转」设为「0°」</p>
      <p>5. 缩放 100%（实际大小），边距设「无」</p>
    </div>

    <div class="label-wrap">
      <div class="label-80x50" id="print-area">
        <div class="label-header">
          <span class="label-unit">中共伊宁县委宣传部</span>
          <span v-if="doc.need_return === 1" class="label-return">
            {{ doc.returned === 1 ? '已退回' : '需退回' }}
          </span>
        </div>
        <div class="label-line">
          <span class="label-key">文件编号</span>
          <span class="label-val">{{ doc.doc_no || '—' }}</span>
        </div>
        <div class="label-line">
          <span class="label-key">收文编号</span>
          <span class="label-val">{{ doc.receive_no || '—' }}</span>
        </div>
        <div class="label-line">
          <span class="label-key">收文日期</span>
          <span class="label-val">{{ doc.received_date || '—' }}</span>
          <span class="label-key" style="margin-left:6px">密级</span>
          <span class="label-val" :class="{ secret: doc.secret_level !== '普通' }">{{ doc.secret_level || '普通' }}</span>
        </div>
        <div class="label-line">
          <span class="label-key">来文单位</span>
          <span class="label-val">{{ doc.from_unit || '—' }}</span>
        </div>
        <div class="label-title">{{ doc.title || '' }}</div>
        <div class="label-platform">伊宁县委宣传部部务工作平台 V1.2</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '../utils/request'

const route = useRoute()
const router = useRouter()
const doc = ref({})

const goBack = () => router.push('/incoming')

const printPage = () => {
  // 触发系统打印对话框（chrome 直接出系统打印框）
  window.print()
}

onMounted(async () => {
  // 优先用 sessionStorage（列表页传来的完整数据）
  const cached = sessionStorage.getItem('printLabel')
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
  max-width: 400px;
  margin: 0 auto 12px;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgba(0,0,0,.1);
  display: flex;
  gap: 8px;
  align-items: center;
}
.toolbar-tip {
  color: #909399;
  font-size: 12px;
}
.print-tips {
  width: 400px;
  margin: 0 auto 12px;
  background: #fdf6ec;
  border: 1px solid #faecd8;
  color: #b88230;
  border-radius: 4px;
  padding: 8px 12px;
  font-size: 12px;
  line-height: 1.8;
}
.label-wrap {
  width: 400px;
  margin: 0 auto;
}
/* 80mm 宽 x 50mm 高 = 302px x 189px (1mm ≈ 3.78px) */
.label-80x50 {
  width: 302px;
  height: 189px;
  background: #fff;
  border: 1px solid #000;
  box-sizing: border-box;
  padding: 8px;
  display: flex;
  flex-direction: column;
  font-family: 'Microsoft YaHei', sans-serif;
}
.label-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #000;
  padding-bottom: 3px;
  margin-bottom: 4px;
}
.label-unit {
  font-size: 11px;
  font-weight: 700;
}
.label-return {
  font-size: 14px;
  font-weight: 700;
  color: #000;
  border: 2px solid #000;
  padding: 0 6px;
  letter-spacing: 2px;
}
.label-line {
  display: flex;
  align-items: center;
  font-size: 10px;
  line-height: 1.6;
}
.label-key {
  color: #666;
  flex-shrink: 0;
}
.label-val {
  font-weight: 600;
  margin-left: 4px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.label-val.secret {
  color: #c8102e;
}
.label-title {
  margin-top: 2px;
  font-size: 10px;
  font-weight: 600;
  color: #303133;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.label-platform {
  margin-top: auto;
  padding-top: 3px;
  border-top: 1px dashed #999;
  font-size: 8px;
  color: #666;
  text-align: center;
  line-height: 1.4;
}
@media print {
  @page {
    size: 80mm 50mm;
    margin: 0;
  }
  body {
    background: #fff !important;
    margin: 0;
    padding: 0;
  }
  .no-print {
    display: none !important;
  }
  .label-wrap {
    width: 80mm;
    height: 50mm;
    margin: 0;
    padding: 0;
    overflow: hidden;
  }
  .label-80x50 {
    width: 80mm;
    height: 50mm;
    border: 1px solid #000;
    box-sizing: border-box;
  }
}
</style>
