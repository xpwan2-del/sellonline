<template>
  <div class="theme-personal-card">
    <div class="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <h3 class="text-lg font-bold theme-text-primary">{{ sectionTitle }}</h3>
        <p class="mt-1 text-sm theme-text-muted">{{ t('personalCenter.affiliate.report.subtitle') }}</p>
      </div>
      <button
        type="button"
        :disabled="summaryLoading || rowsLoading"
        class="inline-flex items-center justify-center rounded-lg border theme-btn-secondary px-3 py-1.5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-60"
        @click="reload"
      >
        {{ t('personalCenter.affiliate.report.refresh') }}
      </button>
    </div>

    <div class="grid grid-cols-2 gap-2 md:grid-cols-4">
      <button
        v-for="range in quickRanges"
        :key="range.key"
        type="button"
        class="rounded-lg border theme-btn-secondary px-3 py-2 text-sm font-semibold"
        @click="applyQuickRange(range.key)"
      >
        {{ range.label }}
      </button>
    </div>

    <div class="mt-3 grid grid-cols-1 gap-3 md:grid-cols-4">
      <input v-model="filters.startDate" type="date" class="form-input-lg" aria-label="开始日期" />
      <input v-model="filters.endDate" type="date" class="form-input-lg" aria-label="结束日期" />
      <select v-model="filters.status" class="form-input-lg">
        <option value="__all__">全部状态</option>
        <option :value="AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM">待确认</option>
        <option :value="AFFILIATE_COMMISSION_STATUS_AVAILABLE">可提现</option>
        <option :value="AFFILIATE_COMMISSION_STATUS_WITHDRAWN">已提现</option>
        <option :value="AFFILIATE_COMMISSION_STATUS_REJECTED">已失效</option>
      </select>
      <button type="button" class="rounded-xl theme-btn-primary px-4 py-2 text-sm font-bold" @click="reload">
        {{ t('personalCenter.affiliate.report.search') }}
      </button>
    </div>

    <div v-if="mode === 'overview'" class="mt-5 grid grid-cols-2 gap-3 md:grid-cols-4">
      <div v-for="item in kpis" :key="item.label" class="rounded-xl border theme-surface-soft p-3">
        <div class="text-xs theme-text-muted">{{ item.label }}</div>
        <div class="mt-2 font-mono text-lg font-bold theme-text-primary">{{ item.value }}</div>
      </div>
    </div>

    <div v-if="mode === 'overview'" class="mt-5 rounded-xl border theme-surface-soft p-4">
      <div class="mb-3 text-sm font-bold theme-text-primary">佣金趋势</div>
      <div v-if="summaryLoading" class="h-40 animate-pulse rounded-lg theme-surface-muted" />
      <div v-else-if="trend.length === 0" class="flex h-40 items-center justify-center text-sm theme-text-muted">暂无趋势数据</div>
      <div v-else class="flex h-40 items-end gap-2 overflow-x-auto border-b border-l border-gray-200/70 px-2 dark:border-white/10">
        <div v-for="item in trend" :key="item.date" class="flex min-w-12 flex-1 flex-col items-center gap-1">
          <span class="font-mono text-[11px] theme-text-muted">{{ item.commission_amount }}</span>
          <div
            class="w-full rounded-t bg-emerald-500"
            :style="{ height: `${Math.max(8, (toNumber(item.commission_amount) / maxTrend) * 92)}px` }"
          />
          <span class="pb-1 text-[11px] theme-text-muted">{{ item.date.slice(5) }}</span>
        </div>
      </div>
    </div>

    <div v-if="mode === 'overview'" class="mt-5 grid gap-3 md:grid-cols-3">
      <div v-for="item in sourceBreakdown" :key="item.source" class="rounded-xl border theme-surface-soft p-3">
        <div class="flex items-center justify-between gap-3 text-sm">
          <span class="theme-text-secondary">{{ item.label }}</span>
          <span class="font-mono theme-text-primary">{{ item.commission_amount }}</span>
        </div>
        <div class="mt-3 h-2 rounded-full bg-gray-200 dark:bg-white/10">
          <div class="h-2 rounded-full bg-sky-500" :style="{ width: `${Math.max(4, (toNumber(item.commission_amount) / maxSource) * 100)}%` }" />
        </div>
        <div class="mt-2 text-xs theme-text-muted">订单金额 {{ item.sales_amount }}</div>
      </div>
    </div>

    <div v-if="mode === 'orders'" class="mt-5">
      <div class="mb-3 flex items-center justify-between">
        <h4 class="text-base font-bold theme-text-primary">{{ t('personalCenter.affiliate.report.downstreamTitle') }}</h4>
        <span class="text-xs theme-text-muted">{{ t('personalCenter.affiliate.report.buyerEmail') }}</span>
      </div>
      <div v-if="rowsLoading" class="space-y-3">
        <div v-for="idx in 3" :key="idx" class="h-24 animate-pulse rounded-xl border theme-surface-muted"></div>
      </div>
      <div v-else-if="rows.length === 0" class="rounded-xl border border-dashed theme-surface-soft px-4 py-6 text-sm theme-text-muted">
        暂无返佣明细
      </div>
      <div v-else class="space-y-3">
        <div v-for="item in rows" :key="item.id" class="rounded-xl border theme-surface-soft p-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <div class="font-mono text-xs theme-text-muted">{{ item.order_no }}</div>
              <div class="mt-1 text-sm theme-text-secondary">买家 {{ item.buyer_email || item.buyer_masked || '-' }}</div>
            </div>
            <div class="text-right">
              <div class="font-mono text-lg font-bold theme-text-primary">{{ item.commission_amount || '0.00' }}</div>
              <span class="theme-badge px-2.5 py-1 text-xs font-semibold" :class="commissionStatusClass(item.status)">
                {{ commissionStatusLabel(item.status) }}
              </span>
            </div>
          </div>
          <div class="mt-3 grid grid-cols-3 gap-2 text-xs">
            <div>
              <div class="theme-text-muted">订单金额</div>
              <div class="mt-1 font-mono theme-text-primary">{{ item.order_total_amount || '0.00' }}</div>
            </div>
            <div>
              <div class="theme-text-muted">返佣比例</div>
              <div class="mt-1 font-mono theme-text-primary">{{ item.rate_percent || '0.00' }}%</div>
            </div>
            <div>
              <div class="theme-text-muted">来源</div>
              <div class="mt-1 theme-text-primary">{{ item.source_label || '-' }}</div>
            </div>
          </div>
          <details v-if="item.commission_items?.length" class="mt-3">
            <summary class="cursor-pointer text-sm theme-text-accent">商品明细</summary>
            <div class="mt-2 space-y-2">
              <div v-for="row in item.commission_items" :key="row.id || `${row.order_item_id}-${row.product_id}`" class="rounded-lg border border-gray-200/70 p-2 text-xs dark:border-white/10">
                <div class="font-medium theme-text-primary">{{ itemTitle(row) }}</div>
                <div class="mt-1 theme-text-muted">{{ itemSku(row) }} · x{{ row.quantity || 0 }}</div>
                <div class="mt-2 grid grid-cols-3 gap-2 font-mono theme-text-secondary">
                  <span>{{ row.base_amount || '0.00' }}</span>
                  <span>{{ row.rate_percent || '0.00' }}%</span>
                  <span>{{ row.commission_amount || '0.00' }}</span>
                </div>
              </div>
            </div>
          </details>
          <div class="mt-3 text-xs theme-text-muted">{{ formatDate(item.created_at) }}</div>
        </div>
      </div>

      <div v-if="pagination.total_page > 1" class="mt-5 flex flex-wrap items-center justify-center gap-3">
        <button
          :disabled="pagination.page <= 1"
          class="rounded-lg border theme-btn-secondary px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-40"
          @click="loadRows(pagination.page - 1)"
        >
          {{ t('personalCenter.affiliate.report.prev') }}
        </button>
        <span class="rounded-full border theme-pill-neutral px-4 py-2 text-sm">
          {{ pagination.page }} / {{ pagination.total_page }}
        </span>
        <button
          :disabled="pagination.page >= pagination.total_page"
          class="rounded-lg border theme-btn-secondary px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-40"
          @click="loadRows(pagination.page + 1)"
        >
          {{ t('personalCenter.affiliate.report.next') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { affiliateAPI } from '../../../api'
import type { AffiliateReportCommissionData, AffiliateReportData } from '../../../api/types'
import {
  AFFILIATE_COMMISSION_STATUS_AVAILABLE,
  AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM,
  AFFILIATE_COMMISSION_STATUS_REJECTED,
  AFFILIATE_COMMISSION_STATUS_WITHDRAWN,
} from '../../../constants/affiliate'
import { useAppStore } from '../../../stores/app'

const props = withDefaults(defineProps<{
  mode?: 'overview' | 'orders'
}>(), {
  mode: 'overview',
})

const { t } = useI18n()
const appStore = useAppStore()
const summaryLoading = ref(false)
const rowsLoading = ref(false)
const report = ref<AffiliateReportData | null>(null)
const rows = ref<AffiliateReportCommissionData[]>([])

const today = new Date()
const yyyyMMdd = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const filters = reactive({
  startDate: yyyyMMdd(new Date(today.getFullYear(), today.getMonth(), today.getDate() - 6)),
  endDate: yyyyMMdd(today),
  status: '__all__',
})

const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0,
  total_page: 1,
})

const quickRanges = [
  { key: 'today', label: '今天' },
  { key: 'week', label: '本周' },
  { key: 'month', label: '本月' },
  { key: 'lastMonth', label: '上月' },
]

const toNumber = (value: unknown) => {
  const parsed = Number(value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

const toApiDate = (value: string, endOfDay = false) => {
  if (!value) return undefined
  const date = new Date(`${value}T00:00:00`)
  if (Number.isNaN(date.getTime())) return undefined
  if (endOfDay) date.setDate(date.getDate() + 1)
  return date.toISOString()
}

const params = () => ({
  start_date: toApiDate(filters.startDate),
  end_date: toApiDate(filters.endDate, true),
  status: filters.status === '__all__' ? undefined : filters.status,
})

const summary = computed(() => report.value?.summary)
const trend = computed(() => report.value?.trend || [])
const sourceBreakdown = computed(() => report.value?.source_breakdown || [])
const maxTrend = computed(() => Math.max(1, ...trend.value.map((item) => toNumber(item.commission_amount))))
const maxSource = computed(() => Math.max(1, ...sourceBreakdown.value.map((item) => toNumber(item.commission_amount))))
const mode = computed(() => props.mode)
const sectionTitle = computed(() => (mode.value === 'orders' ? t('personalCenter.affiliate.report.ordersTitle') : t('personalCenter.affiliate.report.overviewTitle')))

const kpis = computed(() => [
  { label: '总佣金', value: summary.value?.total_commission || '0.00' },
  { label: '可提现', value: summary.value?.available_commission || '0.00' },
  { label: '待确认', value: summary.value?.pending_commission || '0.00' },
  { label: '已提现', value: summary.value?.withdrawn_commission || '0.00' },
  { label: '名下客户', value: String(summary.value?.customer_count || 0) },
  { label: '新客订单', value: String(summary.value?.new_customer_order_count || 0) },
  { label: '复购订单', value: String(summary.value?.repeat_customer_order_count || 0) },
  { label: '有效订单', value: String(summary.value?.valid_order_count || 0) },
  { label: '点击数', value: String(summary.value?.click_count || 0) },
  { label: '转化率', value: `${summary.value?.conversion_rate || '0.00'}%` },
  { label: '平均佣金', value: summary.value?.average_commission || '0.00' },
])

const loadSummary = async () => {
  summaryLoading.value = true
  try {
    const response = await affiliateAPI.reportSummary(params())
    report.value = response.data.data || null
  } catch {
    report.value = null
  } finally {
    summaryLoading.value = false
  }
}

const loadRows = async (page = 1) => {
  rowsLoading.value = true
  try {
    const response = await affiliateAPI.reportCommissions({
      ...params(),
      page,
      page_size: pagination.page_size,
    })
    rows.value = response.data.data || []
    Object.assign(pagination, response.data.pagination || pagination)
  } catch {
    rows.value = []
  } finally {
    rowsLoading.value = false
  }
}

const reload = () => {
  loadSummary()
  loadRows(1)
}

const applyQuickRange = (key: string) => {
  const now = new Date()
  let start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  let end = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  if (key === 'week') {
    const day = start.getDay() || 7
    start.setDate(start.getDate() - day + 1)
  } else if (key === 'month') {
    start = new Date(now.getFullYear(), now.getMonth(), 1)
  } else if (key === 'lastMonth') {
    start = new Date(now.getFullYear(), now.getMonth() - 1, 1)
    end = new Date(now.getFullYear(), now.getMonth(), 0)
  }
  filters.startDate = yyyyMMdd(start)
  filters.endDate = yyyyMMdd(end)
  reload()
}

const itemTitle = (item: NonNullable<AffiliateReportCommissionData['commission_items']>[number]) => {
  const title = item.product_title || {}
  const locale = appStore.locale || 'zh-CN'
  return title[locale] || title['zh-CN'] || title['zh-TW'] || title['en-US'] || `#${item.product_id}`
}

const itemSku = (item: NonNullable<AffiliateReportCommissionData['commission_items']>[number]) => {
  const snapshot = item.sku_snapshot || {}
  const specValues = snapshot.spec_values
  if (specValues && typeof specValues === 'object') {
    const values = Object.values(specValues).map((value) => String(value || '').trim()).filter(Boolean)
    if (values.length) return values.join(' / ')
  }
  return String(snapshot.sku_code || '').trim() || '-'
}

const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

const commissionStatusLabel = (status?: string) => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return '待确认'
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return '可提现'
  if (status === AFFILIATE_COMMISSION_STATUS_REJECTED) return '已失效'
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return '已提现'
  return status || '-'
}

const commissionStatusClass = (status?: string) => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return 'theme-badge-warning'
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return 'theme-badge-success'
  if (status === AFFILIATE_COMMISSION_STATUS_REJECTED) return 'theme-badge-neutral'
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return 'theme-badge-info'
  return 'theme-badge-neutral'
}

onMounted(() => {
  reload()
})
</script>
