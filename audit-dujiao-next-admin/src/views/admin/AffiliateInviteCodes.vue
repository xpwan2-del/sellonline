<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminAffiliateInviteCode } from '@/api/types'
import IdCell from '@/components/IdCell.vue'
import ListPagination from '@/components/ListPagination.vue'
import TableSkeleton from '@/components/TableSkeleton.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { copyText } from '@/utils/clipboard'
import { formatDate } from '@/utils/format'
import { notifyError, notifySuccess } from '@/utils/notify'

const { t } = useI18n()

const INVITE_STATUS_ACTIVE = 'active'
const INVITE_STATUS_DISABLED = 'disabled'
const INVITE_STATUS_USED = 'used'

const loading = ref(true)
const submitting = ref(false)
const operatingId = ref<number | null>(null)
const storefrontBaseURL = ref('')
const rows = ref<AdminAffiliateInviteCode[]>([])
const usageMap = ref<Record<number, AdminAffiliateInviteCode>>({})

const filters = reactive({
  keyword: '',
  status: '__all__',
})

const form = reactive({
  code: '',
  remark: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const pageSizeOptions = [10, 20, 50, 100]
const normalizeSelectValue = (value: string) => (value === '__all__' ? undefined : value)

const fetchRows = async (page = 1) => {
  loading.value = true
  try {
    const response = await adminAPI.getAffiliateInviteCodes({
      page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      status: normalizeSelectValue(filters.status),
    })
    rows.value = (response.data?.data || []) as AdminAffiliateInviteCode[]
    Object.assign(pagination, response.data?.pagination || pagination)
  } catch (err: any) {
    rows.value = []
    notifyError(err?.message || '获取代理邀请码失败')
  } finally {
    loading.value = false
  }
}

const search = () => fetchRows(1)

const changePage = (page: number) => {
  if (page < 1 || page > pagination.total_page) return
  fetchRows(page)
}

const changePageSize = (size: number) => {
  if (size === pagination.page_size) return
  pagination.page_size = size
  fetchRows(1)
}

const createInviteCode = async () => {
  submitting.value = true
  try {
    await adminAPI.createAffiliateInviteCode({
      code: form.code.trim() || undefined,
      remark: form.remark.trim() || undefined,
    })
    form.code = ''
    form.remark = ''
    notifySuccess('代理邀请码已创建')
    await fetchRows(1)
  } catch (err: any) {
    notifyError(err?.message || '创建代理邀请码失败')
  } finally {
    submitting.value = false
  }
}

const updateStatus = async (row: AdminAffiliateInviteCode, status: string) => {
  if (!row?.id || row.status === INVITE_STATUS_USED) return
  operatingId.value = row.id
  try {
    await adminAPI.updateAffiliateInviteCodeStatus(row.id, { status })
    notifySuccess(status === INVITE_STATUS_ACTIVE ? '代理邀请码已启用' : '代理邀请码已停用')
    await fetchRows(pagination.page)
  } catch (err: any) {
    notifyError(err?.message || '更新代理邀请码状态失败')
  } finally {
    operatingId.value = null
  }
}

const loadUsage = async (row: AdminAffiliateInviteCode) => {
  if (!row?.id) return
  try {
    const response = await adminAPI.getAffiliateInviteCodeUsage(row.id)
    usageMap.value = {
      ...usageMap.value,
      [row.id]: response.data?.data as AdminAffiliateInviteCode,
    }
  } catch (err: any) {
    notifyError(err?.message || '获取使用记录失败')
  }
}

const fetchStorefrontBaseURL = async () => {
  try {
    const response = await adminAPI.getSettings({ key: 'site_config' })
    const data = response.data?.data as Record<string, unknown> | undefined
    const brand = data?.brand as Record<string, unknown> | undefined
    storefrontBaseURL.value = String(brand?.site_url || '').trim().replace(/\/+$/, '')
  } catch (err: any) {
    storefrontBaseURL.value = ''
    notifyError(err?.message || '获取站点网址失败')
  }
}

const inviteLink = (row: AdminAffiliateInviteCode) => {
  const code = String(row.code || '').trim()
  if (!storefrontBaseURL.value) return ''
  return `${storefrontBaseURL.value}/auth/register?agent_invite=${encodeURIComponent(code)}`
}

const copyInviteLink = async (row: AdminAffiliateInviteCode) => {
  if (!row?.code) return
  const link = inviteLink(row)
  if (!link) {
    notifyError('请先在系统设置中配置站点网址')
    return
  }
  try {
    await copyText(link)
    notifySuccess(t('admin.affiliatesInviteCodes.copyLinkSuccess'))
  } catch {
    notifyError(t('admin.affiliatesInviteCodes.copyLinkFailed'))
  }
}

const statusLabel = (status?: string) => {
  if (status === INVITE_STATUS_ACTIVE) return '可用'
  if (status === INVITE_STATUS_DISABLED) return '停用'
  if (status === INVITE_STATUS_USED) return '已使用'
  return status || '-'
}

const statusClass = (status?: string) => {
  if (status === INVITE_STATUS_ACTIVE) return 'border-emerald-200 bg-emerald-50 text-emerald-700'
  if (status === INVITE_STATUS_DISABLED) return 'border-zinc-200 bg-zinc-50 text-zinc-700'
  if (status === INVITE_STATUS_USED) return 'border-sky-200 bg-sky-50 text-sky-700'
  return 'border-border bg-muted/30 text-muted-foreground'
}

const usedUserText = (row: AdminAffiliateInviteCode) => {
  const user = usageMap.value[row.id]?.used_by_user || row.used_by_user
  if (!user?.id) return '-'
  return `${user.display_name || user.email || user.id} / #${user.id}`
}

onMounted(() => {
  fetchStorefrontBaseURL()
  fetchRows()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold">代理邀请码</h1>
        <p class="mt-1 text-sm text-muted-foreground">只有使用平台代理邀请码注册的新账号，才会成为一级代理商。</p>
      </div>
      <Button size="sm" variant="outline" @click="fetchRows(pagination.page)">刷新</Button>
    </div>

    <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,220px)_minmax(0,1fr)_auto]" @submit.prevent="createInviteCode">
        <Input v-model="form.code" placeholder="邀请码，可不填" />
        <Input v-model="form.remark" placeholder="备注" />
        <Button type="submit" :disabled="submitting">{{ submitting ? '创建中' : '创建邀请码' }}</Button>
      </form>
    </div>

    <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
      <div class="flex flex-wrap items-center gap-3">
        <div class="w-full md:w-56">
          <Input v-model="filters.keyword" placeholder="邀请码 / 备注" @keyup.enter="search" />
        </div>
        <div class="w-full md:w-44">
          <Select v-model="filters.status" @update:modelValue="search">
            <SelectTrigger class="h-9 w-full">
              <SelectValue placeholder="全部状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__all__">全部状态</SelectItem>
              <SelectItem :value="INVITE_STATUS_ACTIVE">可用</SelectItem>
              <SelectItem :value="INVITE_STATUS_DISABLED">停用</SelectItem>
              <SelectItem :value="INVITE_STATUS_USED">已使用</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button size="sm" @click="search">查询</Button>
      </div>
    </div>

    <div class="overflow-x-auto rounded-xl border border-border bg-card">
      <Table class="min-w-[1080px]">
        <TableHeader class="border-b border-border bg-muted/40 text-xs uppercase text-muted-foreground">
          <TableRow>
            <TableHead class="px-6 py-3">ID</TableHead>
            <TableHead class="px-6 py-3">邀请码</TableHead>
            <TableHead class="px-6 py-3">状态</TableHead>
            <TableHead class="px-6 py-3">使用人</TableHead>
            <TableHead class="px-6 py-3">使用时间</TableHead>
            <TableHead class="px-6 py-3">备注</TableHead>
            <TableHead class="px-6 py-3">创建时间</TableHead>
            <TableHead class="px-6 py-3 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody class="divide-y divide-border">
          <TableRow v-if="loading">
            <TableCell :colspan="8" class="p-0">
              <TableSkeleton :columns="8" :rows="5" />
            </TableCell>
          </TableRow>
          <TableRow v-else-if="rows.length === 0">
            <TableCell colspan="8" class="px-6 py-8 text-center text-muted-foreground">暂无代理邀请码</TableCell>
          </TableRow>
          <TableRow v-for="item in rows" v-else :key="item.id" class="hover:bg-muted/30">
            <TableCell class="px-6 py-4"><IdCell :value="item.id" /></TableCell>
            <TableCell class="px-6 py-4 font-mono text-xs text-foreground">{{ item.code }}</TableCell>
            <TableCell class="px-6 py-4 text-xs">
              <span class="inline-flex rounded-full border px-2.5 py-1 text-xs" :class="statusClass(item.status)">
                {{ statusLabel(item.status) }}
              </span>
            </TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ usedUserText(item) }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ formatDate(item.used_at) }}</TableCell>
            <TableCell class="max-w-[220px] break-words px-6 py-4 text-xs text-muted-foreground">{{ item.remark || '-' }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
            <TableCell class="px-6 py-4 text-right">
              <div class="flex justify-end gap-2">
                <Button size="sm" variant="outline" @click="copyInviteLink(item)">{{ t('admin.affiliatesInviteCodes.copyLink') }}</Button>
                <Button size="sm" variant="outline" @click="loadUsage(item)">使用记录</Button>
                <Button
                  v-if="item.status === INVITE_STATUS_ACTIVE"
                  size="sm"
                  variant="outline"
                  :disabled="operatingId === item.id"
                  @click="updateStatus(item, INVITE_STATUS_DISABLED)"
                >
                  停用
                </Button>
                <Button
                  v-else-if="item.status === INVITE_STATUS_DISABLED"
                  size="sm"
                  variant="outline"
                  :disabled="operatingId === item.id"
                  @click="updateStatus(item, INVITE_STATUS_ACTIVE)"
                >
                  启用
                </Button>
              </div>
            </TableCell>
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
</template>
