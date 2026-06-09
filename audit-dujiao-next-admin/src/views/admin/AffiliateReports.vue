<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { BarChart3, RefreshCw, Search } from 'lucide-vue-next'
import { adminAPI } from '@/api/admin'
import type { AdminProduct, AffiliateReportCommission, AffiliateReportResponse } from '@/api/types'
import {
  AFFILIATE_COMMISSION_STATUS_AVAILABLE,
  AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM,
  AFFILIATE_COMMISSION_STATUS_REJECTED,
  AFFILIATE_COMMISSION_STATUS_WITHDRAWN,
} from '@/constants/affiliate'
import ComplianceGuardWrapper from '@/components/ComplianceGuardWrapper.vue'
import IdCell from '@/components/IdCell.vue'
import ListPagination from '@/components/ListPagination.vue'
import TableSkeleton from '@/components/TableSkeleton.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatDate, getLocalizedText } from '@/utils/format'

const { t } = useI18n()

const loading = ref(true)
const tableLoading = ref(true)
const rows = ref<AffiliateReportCommission[]>([])
const report = ref<AffiliateReportResponse | null>(null)
const productOptions = ref<AdminProduct[]>([])
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

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
  affiliateKeyword: '',
  productId: '__all__',
  status: '__all__',
})

const pageSizeOptions = [10, 20, 50, 100]
const productOptionPageSize = 200
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
  const [year, month, day] = value.split('-').map((part) => Number(part))
  if (!year || !month || !day) return undefined
  const date = new Date(year, month - 1, day, 0, 0, 0)
  if (Number.isNaN(date.getTime())) return undefined
  if (endOfDay) {
    date.setDate(date.getDate() + 1)
  }
  const offsetMinutes = -date.getTimezoneOffset()
  const sign = offsetMinutes >= 0 ? '+' : '-'
  const absOffset = Math.abs(offsetMinutes)
  const offsetHours = String(Math.floor(absOffset / 60)).padStart(2, '0')
  const offsetMins = String(absOffset % 60).padStart(2, '0')
  return `${yyyyMMdd(date)}T00:00:00${sign}${offsetHours}:${offsetMins}`
}

const normalizeSelectValue = (value: string) => (value === '__all__' ? undefined : value)

const params = () => ({
  start_date: toApiDate(filters.startDate),
  end_date: toApiDate(filters.endDate, true),
  affiliate_keyword: filters.affiliateKeyword.trim() || undefined,
  product_id: normalizeSelectValue(filters.productId),
  status: normalizeSelectValue(filters.status),
})

const summary = computed(() => report.value?.summary)
const trend = computed(() => report.value?.trend || [])
const topAffiliates = computed(() => report.value?.top_affiliates || [])
const sourceBreakdown = computed(() => report.value?.source_breakdown || [])
const maxTrend = computed(() => Math.max(1, ...trend.value.map((item) => toNumber(item.commission_amount))))
const maxTop = computed(() => Math.max(1, ...topAffiliates.value.map((item) => toNumber(item.commission_amount))))
const maxSource = computed(() => Math.max(1, ...sourceBreakdown.value.map((item) => toNumber(item.commission_amount))))

const kpis = computed(() => [
  { label: '总销售额', value: summary.value?.total_sales_amount || '0.00' },
  { label: '总佣金', value: summary.value?.total_commission || '0.00' },
  { label: '可提现佣金', value: summary.value?.available_commission || '0.00' },
  { label: '待确认佣金', value: summary.value?.pending_commission || '0.00' },
  { label: '已提现佣金', value: summary.value?.withdrawn_commission || '0.00' },
  { label: '代理客户数', value: String(summary.value?.customer_count || 0) },
  { label: '新客订单', value: String(summary.value?.new_customer_order_count || 0) },
  { label: '复购订单', value: String(summary.value?.repeat_customer_order_count || 0) },
  { label: '有效订单', value: String(summary.value?.valid_order_count || 0) },
  { label: '点击数', value: String(summary.value?.click_count || 0) },
  { label: '转化率', value: `${summary.value?.conversion_rate || '0.00'}%` },
  { label: '平均佣金', value: summary.value?.average_commission || '0.00' },
  { label: '综合返佣率', value: `${summary.value?.average_commission_rate || '0.00'}%` },
])

const fetchSummary = async () => {
  loading.value = true
  try {
    const res = await adminAPI.getAffiliateReportSummary(params())
    report.value = (res.data?.data || null) as AffiliateReportResponse | null
  } catch {
    report.value = null
  } finally {
    loading.value = false
  }
}

const fetchRows = async (page = 1) => {
  tableLoading.value = true
  try {
    const res = await adminAPI.getAffiliateReportCommissions({
      ...params(),
      page,
      page_size: pagination.value.page_size,
    })
    rows.value = (res.data?.data || []) as AffiliateReportCommission[]
    pagination.value = res.data?.pagination || pagination.value
  } catch {
    rows.value = []
  } finally {
    tableLoading.value = false
  }
}

const fetchProductOptions = async () => {
  try {
    const allProducts: AdminProduct[] = []
    let page = 1
    let totalPage = 1
    do {
      const res = await adminAPI.getProducts({ page, page_size: productOptionPageSize })
      allProducts.push(...((res.data?.data || []) as AdminProduct[]))
      totalPage = Number(res.data?.pagination?.total_page || 1)
      page += 1
    } while (page <= totalPage)
    productOptions.value = allProducts
  } catch {
    productOptions.value = []
  }
}

const search = () => {
  fetchSummary()
  fetchRows(1)
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
  search()
}

const changePage = (page: number) => {
  if (page < 1 || page > pagination.value.total_page) return
  fetchRows(page)
}

const changePageSize = (size: number) => {
  if (size === pagination.value.page_size) return
  pagination.value.page_size = size
  fetchRows(1)
}

const statusLabel = (status?: string) => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return t('admin.affiliatesCommissions.status.pendingConfirm')
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return t('admin.affiliatesCommissions.status.available')
  if (status === AFFILIATE_COMMISSION_STATUS_REJECTED) return t('admin.affiliatesCommissions.status.rejected')
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return t('admin.affiliatesCommissions.status.withdrawn')
  return status || '-'
}

const statusClass = (status?: string) => {
  if (status === AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM) return 'border-amber-200 bg-amber-50 text-amber-700'
  if (status === AFFILIATE_COMMISSION_STATUS_AVAILABLE) return 'border-emerald-200 bg-emerald-50 text-emerald-700'
  if (status === AFFILIATE_COMMISSION_STATUS_REJECTED) return 'border-zinc-200 bg-zinc-50 text-zinc-700'
  if (status === AFFILIATE_COMMISSION_STATUS_WITHDRAWN) return 'border-sky-200 bg-sky-50 text-sky-700'
  return 'border-border bg-muted/30 text-muted-foreground'
}

const itemTitle = (item: NonNullable<AffiliateReportCommission['commission_items']>[number]) => {
  return getLocalizedText(item.product_title || {}) || `#${item.product_id}`
}

const itemSku = (item: NonNullable<AffiliateReportCommission['commission_items']>[number]) => {
  const snapshot = item.sku_snapshot || {}
  const specValues = snapshot.spec_values
  if (specValues && typeof specValues === 'object') {
    const values = Object.values(specValues).map((value) => String(value || '').trim()).filter(Boolean)
    if (values.length) return values.join(' / ')
  }
  return String(snapshot.sku_code || '').trim() || '-'
}

onMounted(() => {
  fetchProductOptions()
  search()
})
</script>

<template>
  <ComplianceGuardWrapper>
    <div class="space-y-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold">推广返利 / 返佣统计报表</h1>
          <p class="mt-1 text-sm text-muted-foreground">按支付时间统计返佣，只读展示，不改变订单和佣金数据。</p>
        </div>
        <Button size="sm" variant="outline" :disabled="loading || tableLoading" @click="search">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': loading || tableLoading }" />
          {{ t('admin.common.refresh') }}
        </Button>
      </div>

      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="flex flex-wrap gap-2">
          <Button
            v-for="range in quickRanges"
            :key="range.key"
            size="sm"
            variant="outline"
            @click="applyQuickRange(range.key)"
          >
            {{ range.label }}
          </Button>
        </div>
        <div class="mt-4 grid gap-3 md:grid-cols-5">
          <Input v-model="filters.startDate" type="date" aria-label="开始日期" />
          <Input v-model="filters.endDate" type="date" aria-label="结束日期" />
          <Input v-model="filters.affiliateKeyword" placeholder="推广人邮箱 / 用户名 / 联盟ID" />
          <Select v-model="filters.productId">
            <SelectTrigger class="h-9 w-full">
              <SelectValue placeholder="全部商品" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__all__">全部商品</SelectItem>
              <SelectItem v-for="product in productOptions" :key="product.id" :value="String(product.id)">
                {{ getLocalizedText(product.title) || product.slug || `#${product.id}` }} #{{ product.id }}
              </SelectItem>
            </SelectContent>
          </Select>
          <div class="flex gap-2">
            <Select v-model="filters.status">
              <SelectTrigger class="h-9 w-full">
                <SelectValue placeholder="全部状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">全部状态</SelectItem>
                <SelectItem :value="AFFILIATE_COMMISSION_STATUS_PENDING_CONFIRM">待确认</SelectItem>
                <SelectItem :value="AFFILIATE_COMMISSION_STATUS_AVAILABLE">可提现</SelectItem>
                <SelectItem :value="AFFILIATE_COMMISSION_STATUS_WITHDRAWN">已提现</SelectItem>
                <SelectItem :value="AFFILIATE_COMMISSION_STATUS_REJECTED">已失效</SelectItem>
              </SelectContent>
            </Select>
            <Button size="sm" class="shrink-0" @click="search">
              <Search class="mr-2 h-4 w-4" />
              查询
            </Button>
          </div>
        </div>
      </div>

      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        <div v-for="item in kpis" :key="item.label" class="rounded-xl border border-border bg-card p-4 shadow-sm">
          <div class="text-xs text-muted-foreground">{{ item.label }}</div>
          <div class="mt-2 font-mono text-2xl font-semibold text-foreground">{{ item.value }}</div>
        </div>
      </div>

      <div class="grid gap-4 xl:grid-cols-[1.3fr_1fr]">
        <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
          <div class="mb-4 flex items-center gap-2 text-sm font-medium">
            <BarChart3 class="h-4 w-4 text-primary" />
            佣金趋势图
          </div>
          <div v-if="loading" class="h-56 animate-pulse rounded-md bg-muted" />
          <div v-else-if="trend.length === 0" class="flex h-56 items-center justify-center text-sm text-muted-foreground">暂无趋势数据</div>
          <div v-else class="flex h-56 items-end gap-3 overflow-x-auto border-b border-l border-border px-3 pt-4">
            <div v-for="item in trend" :key="item.date" class="flex min-w-14 flex-1 flex-col items-center gap-2">
              <div class="text-xs font-mono text-muted-foreground">{{ item.commission_amount }}</div>
              <div
                class="w-full rounded-t bg-primary/80"
                :style="{ height: `${Math.max(8, (toNumber(item.commission_amount) / maxTrend) * 150)}px` }"
              />
              <div class="pb-2 text-xs text-muted-foreground">{{ item.date.slice(5) }}</div>
            </div>
          </div>
        </div>

        <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
          <div class="mb-4 text-sm font-medium">推广人排行图</div>
          <div v-if="loading" class="space-y-3">
            <div v-for="i in 5" :key="i" class="h-8 animate-pulse rounded bg-muted" />
          </div>
          <div v-else-if="topAffiliates.length === 0" class="flex h-56 items-center justify-center text-sm text-muted-foreground">暂无排行数据</div>
          <div v-else class="space-y-3">
            <div v-for="item in topAffiliates" :key="item.affiliate_profile_id" class="space-y-1">
              <div class="flex items-center justify-between gap-3 text-xs">
                <span class="min-w-0 truncate">{{ item.email || item.display_name || item.affiliate_code }}</span>
                <span class="font-mono">{{ item.commission_amount }}</span>
              </div>
              <div class="h-2 rounded bg-muted">
                <div class="h-2 rounded bg-emerald-500" :style="{ width: `${Math.max(4, (toNumber(item.commission_amount) / maxTop) * 100)}%` }" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="mb-4 text-sm font-medium">返佣来源拆分</div>
        <div class="grid gap-3 md:grid-cols-3">
          <div v-for="item in sourceBreakdown" :key="item.source" class="rounded-lg border border-border p-3">
            <div class="flex items-center justify-between gap-3 text-sm">
              <span>{{ item.label }}</span>
              <span class="font-mono">{{ item.commission_amount }}</span>
            </div>
            <div class="mt-3 h-2 rounded bg-muted">
              <div class="h-2 rounded bg-sky-500" :style="{ width: `${Math.max(4, (toNumber(item.commission_amount) / maxSource) * 100)}%` }" />
            </div>
            <div class="mt-2 text-xs text-muted-foreground">销售额 {{ item.sales_amount }}</div>
          </div>
        </div>
      </div>

      <div class="overflow-x-auto rounded-xl border border-border bg-card">
        <Table class="min-w-[1120px]">
          <TableHeader class="border-b border-border bg-muted/40 text-xs uppercase text-muted-foreground">
            <TableRow>
              <TableHead class="px-6 py-3">ID</TableHead>
              <TableHead class="min-w-[170px] px-6 py-3">推广人</TableHead>
              <TableHead class="min-w-[160px] px-6 py-3">订单号</TableHead>
              <TableHead class="px-6 py-3">买家</TableHead>
              <TableHead class="px-6 py-3">订单金额</TableHead>
              <TableHead class="px-6 py-3">佣金基数</TableHead>
              <TableHead class="px-6 py-3">比例</TableHead>
              <TableHead class="px-6 py-3">佣金</TableHead>
              <TableHead class="px-6 py-3">来源</TableHead>
              <TableHead class="px-6 py-3">状态</TableHead>
              <TableHead class="min-w-[140px] px-6 py-3">时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody class="divide-y divide-border">
            <TableRow v-if="tableLoading">
              <TableCell colspan="11" class="p-0">
                <TableSkeleton :columns="11" :rows="5" />
              </TableCell>
            </TableRow>
            <TableRow v-else-if="rows.length === 0">
              <TableCell colspan="11" class="px-6 py-8 text-center text-muted-foreground">暂无佣金明细</TableCell>
            </TableRow>
            <TableRow v-for="item in rows" :key="item.id" class="hover:bg-muted/30">
              <TableCell class="px-6 py-4"><IdCell :value="item.id" /></TableCell>
              <TableCell class="min-w-[170px] px-6 py-4 text-xs">
                <div class="text-foreground">{{ item.promoter_name || '-' }}</div>
                <div class="mt-0.5 break-all text-muted-foreground">{{ item.promoter_email || '-' }}</div>
                <div class="mt-0.5 font-mono text-muted-foreground">#{{ item.affiliate_profile_id || '-' }} / {{ item.affiliate_code || '-' }}</div>
              </TableCell>
              <TableCell class="min-w-[160px] px-6 py-4 font-mono text-xs">{{ item.order_no }}</TableCell>
              <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ item.buyer_email || item.buyer_masked || '-' }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ item.order_total_amount || '0.00' }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ item.base_amount || '0.00' }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ item.rate_percent || '0.00' }}%</TableCell>
              <TableCell class="px-6 py-4 text-xs">
                <div class="font-mono">{{ item.commission_amount || '0.00' }}</div>
                <details v-if="item.commission_items?.length" class="mt-2">
                  <summary class="cursor-pointer text-primary">商品明细</summary>
                  <div class="mt-2 min-w-[520px] rounded-md border border-border bg-muted/20 p-2">
                    <div
                      v-for="row in item.commission_items"
                      :key="row.id || `${row.order_item_id}-${row.product_id}`"
                      class="grid grid-cols-[1.5fr_0.8fr_0.5fr_0.8fr_0.8fr_0.8fr_0.9fr] gap-2 py-1 text-[11px]"
                    >
                      <span class="break-words">{{ itemTitle(row) }}</span>
                      <span class="text-muted-foreground">{{ itemSku(row) }}</span>
                      <span class="font-mono">x{{ row.quantity || 0 }}</span>
                      <span class="font-mono">{{ row.base_amount || '0.00' }}</span>
                      <span class="font-mono">{{ row.rate_percent || '0.00' }}%</span>
                      <span class="font-mono">{{ row.commission_amount || '0.00' }}</span>
                      <span>{{ row.source_label || '-' }}</span>
                    </div>
                  </div>
                </details>
              </TableCell>
              <TableCell class="px-6 py-4 text-xs">{{ item.source_label || '-' }}</TableCell>
              <TableCell class="px-6 py-4 text-xs">
                <span class="inline-flex rounded-full border px-2.5 py-1" :class="statusClass(item.status)">
                  {{ statusLabel(item.status) }}
                </span>
              </TableCell>
              <TableCell class="min-w-[140px] px-6 py-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <ListPagination
          :page="pagination.page"
          :total-page="pagination.total_page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          :page-size-options="pageSizeOptions"
          @change-page="changePage"
          @change-page-size="changePageSize"
        />
      </div>
    </div>
  </ComplianceGuardWrapper>
</template>
