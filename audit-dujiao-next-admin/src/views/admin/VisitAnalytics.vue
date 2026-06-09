<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RefreshCw, Search } from 'lucide-vue-next'
import { adminAPI } from '@/api/admin'
import type {
  AdminSiteVisitPageRank,
  AdminSiteVisitRecent,
  AdminSiteVisitSource,
  AdminSiteVisitSummary,
  AdminSiteVisitTrend,
  AdminSiteVisitTrendPoint,
} from '@/api/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import TableSkeleton from '@/components/TableSkeleton.vue'
import ListPagination from '@/components/ListPagination.vue'

type RangeKey = 'today' | 'week' | 'month' | 'last_month' | 'custom'

const rangeOptions: Array<{ value: RangeKey; label: string }> = [
  { value: 'today', label: '今天' },
  { value: 'week', label: '本周' },
  { value: 'month', label: '本月' },
  { value: 'last_month', label: '上月' },
  { value: 'custom', label: '自定义' },
]

const sourceOptions = [
  { value: '__all__', label: '全部来源' },
  { value: 'direct', label: '直接访问' },
  { value: 'affiliate', label: '推广进站' },
  { value: 'search', label: '搜索引擎' },
  { value: 'external', label: '外部链接' },
  { value: 'unknown', label: '未知来源' },
]

const deviceOptions = [
  { value: '__all__', label: '全部设备' },
  { value: 'mobile', label: '移动端' },
  { value: 'desktop', label: 'PC' },
  { value: 'tablet', label: '平板' },
  { value: 'unknown', label: '未知设备' },
]

const filters = reactive({
  range: 'week' as RangeKey,
  startDate: '',
  endDate: '',
  sourceType: '__all__',
  deviceType: '__all__',
})

const loading = ref(false)
const error = ref('')
const summary = ref<AdminSiteVisitSummary | null>(null)
const trends = ref<AdminSiteVisitTrend | null>(null)
const sources = ref<AdminSiteVisitSource[]>([])
const pages = ref<AdminSiteVisitPageRank[]>([])
const recent = ref<AdminSiteVisitRecent[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const totalPage = ref(1)

const queryParams = computed(() => {
  const params: Record<string, unknown> = {
    range: filters.range,
    tz: Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai',
    source_type: filters.sourceType === '__all__' ? undefined : filters.sourceType,
    device_type: filters.deviceType === '__all__' ? undefined : filters.deviceType,
  }
  if (filters.range === 'custom') {
    params.start_date = filters.startDate || undefined
    params.end_date = filters.endDate || undefined
  }
  return params
})

const kpiCards = computed(() => {
  const kpi = summary.value?.kpi
  return [
    { label: 'PV', value: kpi?.pv ?? 0, hint: '访问次数' },
    { label: 'UV', value: kpi?.uv ?? 0, hint: '独立访客' },
    { label: '独立 IP', value: kpi?.unique_ip ?? 0, hint: '按 IP 去重' },
    { label: '推广进站', value: kpi?.affiliate_pv ?? 0, hint: `${kpi?.affiliate_rate ?? '0.00'}%` },
    { label: '游客访问', value: kpi?.visitor_pv ?? 0, hint: '未登录快照' },
    { label: '登录用户访问', value: kpi?.logged_in_pv ?? 0, hint: `${kpi?.logged_in_rate ?? '0.00'}%` },
    { label: '移动端', value: kpi?.mobile_pv ?? 0, hint: `${kpi?.mobile_rate ?? '0.00'}%` },
    { label: 'PC', value: kpi?.desktop_pv ?? 0, hint: '桌面浏览器' },
  ]
})

const maxTrendPV = computed(() => Math.max(1, ...(trends.value?.points || []).map((item) => item.pv)))
const maxSourcePV = computed(() => Math.max(1, ...sources.value.map((item) => item.pv)))
const maxPagePV = computed(() => Math.max(1, ...pages.value.map((item) => item.pv)))

const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

const formatDay = (raw: string) => {
  const parts = String(raw || '').split('-')
  if (parts.length === 3) {
    return `${parts[1]}/${parts[2]}`
  }
  return raw
}

const sourceLabel = (source: string) => {
  const found = sourceOptions.find((item) => item.value === source)
  return found?.label || source || '未知来源'
}

const deviceLabel = (device: string) => {
  const found = deviceOptions.find((item) => item.value === device)
  return found?.label || device || '未知设备'
}

const maskIP = (ip: string) => {
  if (!ip) return '-'
  if (ip.includes(':')) {
    return `${ip.split(':').slice(0, 2).join(':')}:****`
  }
  const parts = ip.split('.')
  if (parts.length === 4) {
    return `${parts[0]}.${parts[1]}.${parts[2]}.*`
  }
  return ip
}

const shortVisitor = (value: string) => {
  if (!value) return '-'
  if (value.length <= 10) return value
  return `${value.slice(0, 6)}...${value.slice(-4)}`
}

const barHeight = (point: AdminSiteVisitTrendPoint, key: keyof AdminSiteVisitTrendPoint) => {
  const raw = Number(point[key] || 0)
  return `${Math.max(raw > 0 ? 8 : 0, Math.round((raw / maxTrendPV.value) * 160))}px`
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const params = queryParams.value
    const [summaryRes, trendRes, sourceRes, pageRes, recentRes] = await Promise.all([
      adminAPI.getSiteVisitSummary(params),
      adminAPI.getSiteVisitTrends(params),
      adminAPI.getSiteVisitSources(params),
      adminAPI.getSiteVisitPages(params),
      adminAPI.getSiteVisitRecent({ ...params, page: page.value, page_size: pageSize.value }),
    ])
    summary.value = summaryRes.data?.data || null
    trends.value = trendRes.data?.data || null
    sources.value = sourceRes.data?.data || []
    pages.value = pageRes.data?.data || []
    recent.value = recentRes.data?.data || []
    const pagination = recentRes.data?.pagination
    total.value = Number(pagination?.total || 0)
    totalPage.value = Number(pagination?.total_page || 1)
  } catch (err: any) {
    error.value = err?.message || '获取进站统计失败'
  } finally {
    loading.value = false
  }
}

const applyRange = (range: RangeKey) => {
  filters.range = range
  page.value = 1
  loadData()
}

const handleSearch = () => {
  page.value = 1
  loadData()
}

const changePage = (nextPage: number) => {
  page.value = nextPage
  loadData()
}

const changePageSize = (nextPageSize: number) => {
  pageSize.value = nextPageSize
  page.value = 1
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div>
        <h1 class="text-2xl font-semibold">进站统计</h1>
        <p class="mt-1 text-sm text-muted-foreground">按进站时间统计访问、来源、设备和页面排行。</p>
      </div>
      <Button variant="outline" size="sm" class="h-9 gap-2" :disabled="loading" @click="loadData">
        <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
        刷新
      </Button>
    </div>

    <div class="flex flex-col gap-3 rounded-lg border border-border bg-card p-4 shadow-sm">
      <div class="flex flex-wrap gap-2">
        <Button
          v-for="item in rangeOptions"
          :key="item.value"
          size="sm"
          :variant="filters.range === item.value ? 'default' : 'outline'"
          class="h-8"
          @click="applyRange(item.value)"
        >
          {{ item.label }}
        </Button>
      </div>
      <div class="grid gap-3 md:grid-cols-5">
        <Input v-model="filters.startDate" type="date" class="h-9" :disabled="filters.range !== 'custom'" />
        <Input v-model="filters.endDate" type="date" class="h-9" :disabled="filters.range !== 'custom'" />
        <Select v-model="filters.sourceType">
          <SelectTrigger class="h-9">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="item in sourceOptions" :key="item.value" :value="item.value">{{ item.label }}</SelectItem>
          </SelectContent>
        </Select>
        <Select v-model="filters.deviceType">
          <SelectTrigger class="h-9">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="item in deviceOptions" :key="item.value" :value="item.value">{{ item.label }}</SelectItem>
          </SelectContent>
        </Select>
        <Button class="h-9 gap-2" :disabled="loading" @click="handleSearch">
          <Search class="h-4 w-4" />
          查询
        </Button>
      </div>
    </div>

    <div v-if="error" class="rounded-lg border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {{ error }}
    </div>

    <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <Card v-for="item in kpiCards" :key="item.label" class="rounded-lg border-border shadow-sm">
        <CardHeader class="pb-2">
          <CardTitle class="text-xs font-medium text-muted-foreground">{{ item.label }}</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="font-mono text-2xl font-semibold">{{ item.value }}</div>
          <div class="mt-1 text-xs text-muted-foreground">{{ item.hint }}</div>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-4 xl:grid-cols-[1.4fr_1fr]">
      <Card class="rounded-lg border-border shadow-sm">
        <CardHeader>
          <CardTitle class="text-sm">每天进站柱形图</CardTitle>
        </CardHeader>
        <CardContent>
          <div v-if="loading" class="h-56 animate-pulse rounded bg-muted/60"></div>
          <div v-else-if="!trends || trends.points.length === 0" class="py-12 text-center text-sm text-muted-foreground">暂无趋势数据</div>
          <div v-else class="overflow-x-auto">
            <div class="flex h-64 min-w-[680px] items-end gap-4 border-b border-border px-2 pb-8">
              <div v-for="point in trends.points" :key="point.date" class="flex flex-1 flex-col items-center gap-2">
                <div class="flex h-44 items-end gap-1">
                  <div class="w-3 rounded-t bg-sky-500" :style="{ height: barHeight(point, 'pv') }" :title="`PV ${point.pv}`"></div>
                  <div class="w-3 rounded-t bg-emerald-500" :style="{ height: barHeight(point, 'uv') }" :title="`UV ${point.uv}`"></div>
                </div>
                <div class="font-mono text-xs text-muted-foreground">{{ formatDay(point.date) }}</div>
              </div>
            </div>
            <div class="mt-3 flex items-center gap-4 text-xs text-muted-foreground">
              <span class="inline-flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-sky-500"></span>PV</span>
              <span class="inline-flex items-center gap-1"><span class="h-2 w-2 rounded-full bg-emerald-500"></span>UV</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card class="rounded-lg border-border shadow-sm">
        <CardHeader>
          <CardTitle class="text-sm">来源占比</CardTitle>
        </CardHeader>
        <CardContent>
          <div v-if="loading" class="space-y-3">
            <div v-for="i in 5" :key="i" class="h-8 animate-pulse rounded bg-muted/60"></div>
          </div>
          <div v-else-if="sources.length === 0" class="py-12 text-center text-sm text-muted-foreground">暂无来源数据</div>
          <div v-else class="space-y-4">
            <div v-for="item in sources" :key="item.source_type" class="space-y-1">
              <div class="flex items-center justify-between text-sm">
                <span>{{ sourceLabel(item.source_type) }}</span>
                <span class="font-mono">{{ item.pv }} / {{ item.rate }}%</span>
              </div>
              <div class="h-3 overflow-hidden rounded bg-muted">
                <div class="h-full rounded bg-primary" :style="{ width: `${Math.max(4, (item.pv / maxSourcePV) * 100)}%` }"></div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-4 xl:grid-cols-2">
      <Card class="rounded-lg border-border shadow-sm">
        <CardHeader>
          <CardTitle class="text-sm">页面排行</CardTitle>
        </CardHeader>
        <CardContent>
          <div v-if="pages.length === 0" class="py-10 text-center text-sm text-muted-foreground">暂无页面数据</div>
          <div v-else class="space-y-4">
            <div v-for="item in pages" :key="item.path" class="space-y-1">
              <div class="flex items-center justify-between gap-3 text-sm">
                <span class="truncate font-mono">{{ item.path }}</span>
                <span class="shrink-0 font-mono">{{ item.pv }}</span>
              </div>
              <div class="h-3 overflow-hidden rounded bg-muted">
                <div class="h-full rounded bg-emerald-500" :style="{ width: `${Math.max(4, (item.pv / maxPagePV) * 100)}%` }"></div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card class="rounded-lg border-border shadow-sm">
        <CardHeader>
          <CardTitle class="text-sm">来源拆分</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="rounded-lg border border-border bg-background p-4">
              <div class="text-xs text-muted-foreground">直接访问</div>
              <div class="mt-2 font-mono text-xl font-semibold">{{ summary?.kpi.direct_pv ?? 0 }}</div>
            </div>
            <div class="rounded-lg border border-border bg-background p-4">
              <div class="text-xs text-muted-foreground">推广进站</div>
              <div class="mt-2 font-mono text-xl font-semibold">{{ summary?.kpi.affiliate_pv ?? 0 }}</div>
            </div>
            <div class="rounded-lg border border-border bg-background p-4">
              <div class="text-xs text-muted-foreground">搜索引擎</div>
              <div class="mt-2 font-mono text-xl font-semibold">{{ summary?.kpi.search_pv ?? 0 }}</div>
            </div>
            <div class="rounded-lg border border-border bg-background p-4">
              <div class="text-xs text-muted-foreground">外部链接</div>
              <div class="mt-2 font-mono text-xl font-semibold">{{ summary?.kpi.external_pv ?? 0 }}</div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <div class="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div class="border-b border-border px-6 py-4">
        <h2 class="text-sm font-semibold">最近进站明细</h2>
      </div>
      <div class="overflow-x-auto">
        <Table class="min-w-[980px]">
          <TableHeader class="border-b border-border bg-muted/40 text-xs uppercase text-muted-foreground">
            <TableRow>
              <TableHead class="px-6 py-3">ID</TableHead>
              <TableHead class="px-6 py-3">时间</TableHead>
              <TableHead class="px-6 py-3">访客</TableHead>
              <TableHead class="px-6 py-3">用户 ID</TableHead>
              <TableHead class="px-6 py-3">页面</TableHead>
              <TableHead class="px-6 py-3">来源</TableHead>
              <TableHead class="px-6 py-3">设备</TableHead>
              <TableHead class="px-6 py-3">推广码</TableHead>
              <TableHead class="px-6 py-3">IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody class="divide-y divide-border">
            <TableRow v-if="loading">
              <TableCell :colspan="9" class="p-0">
                <TableSkeleton :columns="9" :rows="5" />
              </TableCell>
            </TableRow>
            <TableRow v-else-if="recent.length === 0">
              <TableCell colspan="9" class="px-6 py-8 text-center text-muted-foreground">暂无进站明细</TableCell>
            </TableRow>
            <TableRow v-for="item in recent" :key="item.id" class="hover:bg-muted/30">
              <TableCell class="px-6 py-4 font-mono">#{{ item.id }}</TableCell>
              <TableCell class="min-w-[150px] px-6 py-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ shortVisitor(item.visitor_key) }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ item.user_id || '-' }}</TableCell>
              <TableCell class="max-w-[260px] truncate px-6 py-4 font-mono text-xs" :title="item.path">{{ item.path }}</TableCell>
              <TableCell class="px-6 py-4 text-xs">{{ sourceLabel(item.source_type) }}</TableCell>
              <TableCell class="px-6 py-4 text-xs">{{ deviceLabel(item.device_type) }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ item.affiliate_code || '-' }}</TableCell>
              <TableCell class="px-6 py-4 font-mono text-xs">{{ maskIP(item.client_ip) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <ListPagination
        :page="page"
        :total-page="totalPage"
        :total="total"
        :page-size="pageSize"
        :page-size-options="[20, 50, 100]"
        @change-page="changePage"
        @change-page-size="changePageSize"
      />
    </div>
  </div>
</template>
