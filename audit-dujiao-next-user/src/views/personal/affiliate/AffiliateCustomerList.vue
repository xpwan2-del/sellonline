<template>
  <div class="theme-personal-card">
    <div class="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div>
        <h3 class="text-lg font-bold theme-text-primary">{{ t('personalCenter.affiliate.customers.title') }}</h3>
        <p class="mt-1 text-sm theme-text-muted">{{ t('personalCenter.affiliate.customers.subtitle') }}</p>
      </div>
      <button
        type="button"
        :disabled="loading"
        class="inline-flex items-center justify-center rounded-lg border theme-btn-secondary px-3 py-1.5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-60"
        @click="loadRows(pagination.page)"
      >
        {{ t('personalCenter.affiliate.customers.refresh') }}
      </button>
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1fr)_auto_auto]">
      <input v-model.trim="filters.keyword" type="text" class="form-input-lg" :placeholder="t('personalCenter.affiliate.customers.keywordPlaceholder')" @keyup.enter="loadRows(1)" />
      <input v-model="filters.registeredFrom" type="date" class="form-input-lg" aria-label="开始注册时间" />
      <input v-model="filters.registeredTo" type="date" class="form-input-lg" aria-label="结束注册时间" />
      <button type="button" class="rounded-xl theme-btn-primary px-4 py-2 text-sm font-bold" @click="loadRows(1)">{{ t('personalCenter.affiliate.customers.search') }}</button>
      <button type="button" class="rounded-xl border theme-btn-secondary px-4 py-2 text-sm font-semibold" @click="resetFilters">{{ t('personalCenter.affiliate.customers.reset') }}</button>
    </div>

    <div v-if="loading" class="mt-5 space-y-3">
      <div v-for="idx in 3" :key="idx" class="h-14 animate-pulse rounded-xl border theme-surface-muted"></div>
    </div>
    <div v-else-if="rows.length === 0" class="mt-5 rounded-xl border border-dashed theme-surface-soft px-4 py-6 text-sm theme-text-muted">
      {{ t('personalCenter.affiliate.customers.empty') }}
    </div>
    <div v-else class="mt-5 overflow-x-auto rounded-xl border border-gray-200/70 dark:border-white/10">
      <table class="min-w-full divide-y divide-gray-200 text-left text-sm dark:divide-white/10">
        <thead class="bg-gray-50/80 text-xs uppercase tracking-wide text-gray-500 dark:bg-white/5 dark:text-gray-400">
          <tr>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.id') }}</th>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.email') }}</th>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.displayName') }}</th>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.registeredAt') }}</th>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.lastOrderAt') }}</th>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.orderCount') }}</th>
            <th class="px-4 py-3 font-semibold">{{ t('personalCenter.affiliate.customers.commissionAmount') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 dark:divide-white/10">
          <tr v-for="item in rows" :key="item.id">
            <td class="px-4 py-3 font-mono text-xs theme-text-primary">#{{ item.id }}</td>
            <td class="px-4 py-3 text-xs theme-text-primary">{{ item.email || '-' }}</td>
            <td class="px-4 py-3 text-xs theme-text-secondary">{{ item.display_name || '-' }}</td>
            <td class="px-4 py-3 text-xs theme-text-muted">{{ formatDate(item.registered_at) }}</td>
            <td class="px-4 py-3 text-xs theme-text-muted">{{ formatDate(item.last_order_at) }}</td>
            <td class="px-4 py-3 font-mono text-xs theme-text-primary">{{ item.order_count || 0 }}</td>
            <td class="px-4 py-3 font-mono text-xs theme-text-primary">{{ item.commission_amount || '0.00' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination.total_page > 1" class="mt-5 flex flex-wrap items-center justify-center gap-3">
      <button
        :disabled="pagination.page <= 1"
        class="rounded-lg border theme-btn-secondary px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-40"
        @click="loadRows(pagination.page - 1)"
      >
        {{ t('personalCenter.affiliate.customers.prev') }}
      </button>
      <span class="rounded-full border theme-pill-neutral px-4 py-2 text-sm">
        {{ pagination.page }} / {{ pagination.total_page }}
      </span>
      <button
        :disabled="pagination.page >= pagination.total_page"
        class="rounded-lg border theme-btn-secondary px-4 py-2 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-40"
        @click="loadRows(pagination.page + 1)"
      >
        {{ t('personalCenter.affiliate.customers.next') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { affiliateAPI } from '../../../api'
import type { AffiliateCustomerData } from '../../../api/types'

const { t } = useI18n()
const loading = ref(false)
const rows = ref<AffiliateCustomerData[]>([])

const filters = reactive({
  keyword: '',
  registeredFrom: '',
  registeredTo: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const yyyyMMdd = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const toApiDate = (value: string, endOfDay = false) => {
  if (!value) return undefined
  const [year, month, day] = value.split('-').map((part) => Number(part))
  if (!year || !month || !day) return undefined
  const date = new Date(year, month - 1, day, 0, 0, 0)
  if (Number.isNaN(date.getTime())) return undefined
  if (endOfDay) date.setDate(date.getDate() + 1)
  const offsetMinutes = -date.getTimezoneOffset()
  const sign = offsetMinutes >= 0 ? '+' : '-'
  const absOffset = Math.abs(offsetMinutes)
  const offsetHours = String(Math.floor(absOffset / 60)).padStart(2, '0')
  const offsetMins = String(absOffset % 60).padStart(2, '0')
  return `${yyyyMMdd(date)}T00:00:00${sign}${offsetHours}:${offsetMins}`
}

const loadRows = async (page = 1) => {
  loading.value = true
  try {
    const response = await affiliateAPI.customers({
      page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      registered_from: toApiDate(filters.registeredFrom),
      registered_to: toApiDate(filters.registeredTo, true),
    })
    rows.value = response.data.data || []
    Object.assign(pagination, response.data.pagination || pagination)
  } catch {
    rows.value = []
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.keyword = ''
  filters.registeredFrom = ''
  filters.registeredTo = ''
  loadRows(1)
}

const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return raw
  return date.toLocaleString()
}

onMounted(() => {
  loadRows()
})
</script>
