<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminAffiliateApplication } from '@/api/types'
import IdCell from '@/components/IdCell.vue'
import ListPagination from '@/components/ListPagination.vue'
import TableSkeleton from '@/components/TableSkeleton.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { formatDate } from '@/utils/format'
import { notifyError, notifySuccess } from '@/utils/notify'

const APPLICATION_STATUS_PENDING = 'pending'
const APPLICATION_STATUS_CONTACTED = 'contacted'
const APPLICATION_STATUS_REJECTED = 'rejected'

const loading = ref(true)
const submitting = ref(false)
const rows = ref<AdminAffiliateApplication[]>([])
const selectedRow = ref<AdminAffiliateApplication | null>(null)
const statusDialogOpen = ref(false)

const filters = reactive({
  keyword: '',
  status: '__all__',
})

const statusForm = reactive({
  status: APPLICATION_STATUS_CONTACTED,
  admin_note: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_page: 1,
})

const pageSizeOptions = [10, 20, 50, 100]
const normalizeSelectValue = (value: string) => (value === '__all__' ? undefined : value)

const selectedUserText = computed(() => {
  if (!selectedRow.value) return '-'
  const user = selectedRow.value.user
  return user?.display_name || user?.email || selectedRow.value.email || `#${selectedRow.value.user_id}`
})

const fetchRows = async (page = 1) => {
  loading.value = true
  try {
    const response = await adminAPI.getAffiliateApplications({
      page,
      page_size: pagination.page_size,
      keyword: filters.keyword.trim() || undefined,
      status: normalizeSelectValue(filters.status),
    })
    rows.value = (response.data?.data || []) as AdminAffiliateApplication[]
    Object.assign(pagination, response.data?.pagination || pagination)
  } catch (err: any) {
    rows.value = []
    notifyError(err?.message || '获取代理申请失败')
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

const openStatusDialog = (row: AdminAffiliateApplication) => {
  selectedRow.value = row
  statusForm.status = row.status === APPLICATION_STATUS_PENDING ? APPLICATION_STATUS_CONTACTED : row.status
  statusForm.admin_note = row.admin_note || ''
  statusDialogOpen.value = true
}

const updateStatus = async () => {
  if (!selectedRow.value?.id) return
  submitting.value = true
  try {
    await adminAPI.updateAffiliateApplicationStatus(selectedRow.value.id, {
      status: statusForm.status,
      admin_note: statusForm.admin_note.trim() || undefined,
    })
    notifySuccess('代理申请状态已更新')
    statusDialogOpen.value = false
    await fetchRows(pagination.page)
  } catch (err: any) {
    notifyError(err?.message || '更新代理申请状态失败')
  } finally {
    submitting.value = false
  }
}

const statusLabel = (status?: string) => {
  if (status === APPLICATION_STATUS_PENDING) return '待处理'
  if (status === APPLICATION_STATUS_CONTACTED) return '已联系'
  if (status === APPLICATION_STATUS_REJECTED) return '已拒绝'
  return status || '-'
}

const statusClass = (status?: string) => {
  if (status === APPLICATION_STATUS_PENDING) return 'border-amber-200 bg-amber-50 text-amber-700'
  if (status === APPLICATION_STATUS_CONTACTED) return 'border-sky-200 bg-sky-50 text-sky-700'
  if (status === APPLICATION_STATUS_REJECTED) return 'border-zinc-200 bg-zinc-50 text-zinc-700'
  return 'border-border bg-muted/30 text-muted-foreground'
}

const userText = (row: AdminAffiliateApplication) => {
  const user = row.user
  const name = user?.display_name || user?.email || row.email || '-'
  return `${name} / #${row.user_id}`
}

onMounted(() => {
  fetchRows()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold">代理申请</h1>
        <p class="mt-1 text-sm text-muted-foreground">普通用户提交申请后在这里查看和标记处理状态，不会自动开通一级代理商。</p>
      </div>
      <Button size="sm" variant="outline" @click="fetchRows(pagination.page)">刷新</Button>
    </div>

    <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
      <div class="flex flex-wrap items-center gap-3">
        <div class="w-full md:w-72">
          <Input v-model="filters.keyword" placeholder="用户邮箱 / 电话 / 申请说明" @keyup.enter="search" />
        </div>
        <div class="w-full md:w-44">
          <Select v-model="filters.status" @update:modelValue="search">
            <SelectTrigger class="h-9 w-full">
              <SelectValue placeholder="全部状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__all__">全部状态</SelectItem>
              <SelectItem :value="APPLICATION_STATUS_PENDING">待处理</SelectItem>
              <SelectItem :value="APPLICATION_STATUS_CONTACTED">已联系</SelectItem>
              <SelectItem :value="APPLICATION_STATUS_REJECTED">已拒绝</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button size="sm" @click="search">查询</Button>
      </div>
    </div>

    <div class="overflow-x-auto rounded-xl border border-border bg-card">
      <Table class="min-w-[1120px]">
        <TableHeader class="border-b border-border bg-muted/40 text-xs uppercase text-muted-foreground">
          <TableRow>
            <TableHead class="px-6 py-3">ID</TableHead>
            <TableHead class="px-6 py-3">用户</TableHead>
            <TableHead class="px-6 py-3">邮箱</TableHead>
            <TableHead class="px-6 py-3">电话</TableHead>
            <TableHead class="px-6 py-3">状态</TableHead>
            <TableHead class="px-6 py-3">申请说明</TableHead>
            <TableHead class="px-6 py-3">后台备注</TableHead>
            <TableHead class="px-6 py-3">申请时间</TableHead>
            <TableHead class="px-6 py-3 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody class="divide-y divide-border">
          <TableRow v-if="loading">
            <TableCell :colspan="9" class="p-0">
              <TableSkeleton :columns="9" :rows="5" />
            </TableCell>
          </TableRow>
          <TableRow v-else-if="rows.length === 0">
            <TableCell colspan="9" class="px-6 py-8 text-center text-muted-foreground">暂无代理申请</TableCell>
          </TableRow>
          <TableRow v-for="item in rows" v-else :key="item.id" class="hover:bg-muted/30">
            <TableCell class="px-6 py-4"><IdCell :value="item.id" /></TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ userText(item) }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-foreground">{{ item.email || '-' }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-foreground">{{ item.phone || '-' }}</TableCell>
            <TableCell class="px-6 py-4 text-xs">
              <span class="inline-flex rounded-full border px-2.5 py-1 text-xs" :class="statusClass(item.status)">
                {{ statusLabel(item.status) }}
              </span>
            </TableCell>
            <TableCell class="max-w-[220px] break-words px-6 py-4 text-xs text-muted-foreground">{{ item.message || '-' }}</TableCell>
            <TableCell class="max-w-[220px] break-words px-6 py-4 text-xs text-muted-foreground">{{ item.admin_note || '-' }}</TableCell>
            <TableCell class="px-6 py-4 text-xs text-muted-foreground">{{ formatDate(item.created_at) }}</TableCell>
            <TableCell class="px-6 py-4 text-right">
              <Button size="sm" variant="outline" @click="openStatusDialog(item)">处理</Button>
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

    <Dialog v-model:open="statusDialogOpen">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>处理代理申请</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div class="rounded-lg border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
            {{ selectedUserText }}
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium">状态</label>
            <Select v-model="statusForm.status">
              <SelectTrigger class="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="APPLICATION_STATUS_PENDING">待处理</SelectItem>
                <SelectItem :value="APPLICATION_STATUS_CONTACTED">已联系</SelectItem>
                <SelectItem :value="APPLICATION_STATUS_REJECTED">已拒绝</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium">后台备注</label>
            <Textarea v-model="statusForm.admin_note" rows="4" placeholder="填写处理记录或备注" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="statusDialogOpen = false">取消</Button>
          <Button :disabled="submitting" @click="updateStatus">{{ submitting ? '保存中' : '保存' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
