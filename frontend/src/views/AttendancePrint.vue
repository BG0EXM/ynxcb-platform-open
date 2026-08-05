<template>
  <div>
    <div class="no-print toolbar">
      <el-radio-group v-model="viewMode">
        <el-radio-button value="monthly">月度统计</el-radio-button>
        <el-radio-button value="yearly">年度统计</el-radio-button>
      </el-radio-group>
      <el-date-picker v-model="period" :type="viewMode === 'monthly' ? 'month' : 'year'"
        :value-format="viewMode === 'monthly' ? 'YYYY-MM' : 'YYYY'" placeholder="选择统计期间"
        style="width:150px" @change="loadData" />
      <el-button type="primary" :icon="'Printer'" @click="printPage">一键打印</el-button>
      <el-button :icon="'Back'" @click="goBack">返回</el-button>
    </div>

    <div class="a4-page" id="print-area">
      <div class="doc-header">
        <span class="unit-name">中共伊宁县委宣传部</span>
      </div>
      <div class="doc-title">{{ title }}</div>

      <!-- 月度统计 -->
      <template v-if="viewMode === 'monthly' && monthly">
        <table class="stat-table">
          <thead>
            <tr>
              <th style="width:50px">序号</th>
              <th style="width:90px">姓名</th>
              <th style="width:100px">部门</th>
              <th>出勤</th>
              <th>请假</th>
              <th>出差</th>
              <th>未到</th>
              <th>年假</th>
              <th>病假</th>
              <th>事假</th>
              <th>其他</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(r, i) in monthly.list" :key="r.user_id">
              <td class="center">{{ i + 1 }}</td>
              <td class="center">{{ r.user_name }}</td>
              <td class="center">{{ r.department }}</td>
              <td class="center">{{ r.present }}</td>
              <td class="center">{{ r.leave }}</td>
              <td class="center">{{ r.trip }}</td>
              <td class="center">{{ r.absent }}</td>
              <td class="center">{{ r.annual_days }}</td>
              <td class="center">{{ r.sick_days }}</td>
              <td class="center">{{ r.personal_days }}</td>
              <td class="center">{{ r.other_days }}</td>
            </tr>
            <tr v-if="!monthly.list || !monthly.list.length">
              <td colspan="11" class="center empty">暂无考勤数据</td>
            </tr>
          </tbody>
          <tfoot v-if="monthly.total">
            <tr>
              <td class="center" colspan="2"><b>合计</b></td>
              <td></td>
              <td class="center"><b>{{ monthly.total.present }}</b></td>
              <td class="center"><b>{{ monthly.total.leave }}</b></td>
              <td class="center"><b>{{ monthly.total.trip }}</b></td>
              <td class="center"><b>{{ monthly.total.absent }}</b></td>
              <td colspan="4"></td>
            </tr>
          </tfoot>
        </table>
      </template>

      <!-- 年度统计 -->
      <template v-if="viewMode === 'yearly' && yearly">
        <div class="year-title">按月汇总</div>
        <table class="stat-table">
          <thead>
            <tr>
              <th style="width:80px">月份</th>
              <th>出勤</th>
              <th>请假</th>
              <th>出差</th>
              <th>未到</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in yearly.monthly" :key="r.month">
              <td class="center">{{ r.month }}</td>
              <td class="center">{{ r.present }}</td>
              <td class="center">{{ r.leave }}</td>
              <td class="center">{{ r.trip }}</td>
              <td class="center">{{ r.absent }}</td>
            </tr>
            <tr v-if="!yearly.monthly || !yearly.monthly.length">
              <td colspan="5" class="center empty">暂无考勤数据</td>
            </tr>
          </tbody>
          <tfoot v-if="yearly.total">
            <tr>
              <td class="center"><b>合计</b></td>
              <td class="center"><b>{{ yearly.total.present }}</b></td>
              <td class="center"><b>{{ yearly.total.leave }}</b></td>
              <td class="center"><b>{{ yearly.total.trip }}</b></td>
              <td class="center"><b>{{ yearly.total.absent }}</b></td>
            </tr>
          </tfoot>
        </table>

        <div class="year-title mt-16">年度请假汇总（天）</div>
        <table class="stat-table" v-if="yearly.leave">
          <thead>
            <tr>
              <th>年假</th>
              <th>病假</th>
              <th>事假</th>
              <th>其他</th>
              <th>总计天数</th>
              <th>总计次数</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td class="center">{{ yearly.leave.annual }}</td>
              <td class="center">{{ yearly.leave.sick }}</td>
              <td class="center">{{ yearly.leave.personal }}</td>
              <td class="center">{{ yearly.leave.other }}</td>
              <td class="center"><b>{{ yearly.leave.total_days }}</b></td>
              <td class="center"><b>{{ yearly.leave.total_count }}</b></td>
            </tr>
          </tbody>
        </table>
      </template>

      <div class="doc-footer">
        <span>统计期间：{{ period }} · 打印日期：{{ today }}</span>
        <span class="footer-right">
          <div>中共伊宁县委宣传部办公室</div>
          <div>伊宁县委宣传部部务工作平台V1.2</div>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '../utils/request'
import dayjs from 'dayjs'

const router = useRouter()
const viewMode = ref('monthly')
const period = ref(dayjs().format('YYYY-MM'))
const monthly = ref(null)
const yearly = ref(null)

const today = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
})

const title = computed(() => {
  if (viewMode.value === 'monthly') return `${period.value} 月度考勤统计`
  return `${period.value} 年度考勤统计`
})

const goBack = () => router.push('/attendance')

const printPage = () => window.print()

const loadData = async () => {
  try {
    if (viewMode.value === 'monthly') {
      const res = await request.get('/attendance/monthly', { params: { month: period.value } })
      monthly.value = res
    } else {
      const res = await request.get('/attendance/yearly', { params: { year: period.value } })
      yearly.value = res
    }
  } catch (e) {}
}

watch(viewMode, (v) => {
  period.value = v === 'monthly' ? dayjs().format('YYYY-MM') : dayjs().format('YYYY')
  loadData()
})

onMounted(loadData)
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
  align-items: center;
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
  font-size: 22px;
  font-weight: 700;
  margin: 20px 0 20px;
  letter-spacing: 3px;
}
.year-title {
  font-size: 15px;
  font-weight: 600;
  margin: 16px 0 8px;
  color: #303133;
}
.stat-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #000;
}
.stat-table th,
.stat-table td {
  border: 1px solid #000;
  padding: 7px 8px;
  font-size: 13px;
}
.stat-table th {
  background: #f5f5f5;
  font-weight: 600;
}
.center {
  text-align: center;
}
.empty {
  color: #999;
}
.mt-16 {
  margin-top: 16px;
}
.doc-footer {
  position: absolute;
  bottom: 36px;
  left: 56px;
  right: 56px;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  font-size: 12px;
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
