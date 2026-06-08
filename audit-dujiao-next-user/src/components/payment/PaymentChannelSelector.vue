<template>
  <div v-if="props.channels.length > 0" class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-3">
    <button v-for="channel in props.channels" :key="channel.id"
      :disabled="isDisabled(channel)"
      :title="isDisabled(channel) ? channelHint(channel) : ''"
      @click="handleSelect(channel)"
      class="h-14 rounded-lg border px-4 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50"
      :class="channelButtonClass(channel)">
      <div class="flex h-full items-center justify-between gap-3">
        <div class="flex min-w-0 items-center gap-2.5">
          <img v-if="channel.icon" :src="getImageUrl(channel.icon)" loading="lazy" class="h-7 w-7 rounded object-contain shrink-0" />
          <span v-else class="flex h-7 w-7 shrink-0 items-center justify-center rounded text-base font-black text-white"
            :class="channelIconClass(channel)">
            {{ channelIconText(channel) }}
          </span>
          <div class="min-w-0 truncate text-sm font-medium" :class="channelTextClass(channel)">
            {{ channel.name }}<span v-if="channelFeeLabel(channel)" class="ml-1">{{ channelFeeLabel(channel) }}</span>
          </div>
        </div>
        <span v-if="props.modelValue === channel.id && !isDisabled(channel)"
          class="h-2.5 w-2.5 shrink-0 rounded-full"
          :class="selectedDotClass(channel)"
          aria-hidden="true"></span>
      </div>
      <div v-if="isDisabled(channel)" class="mt-2 text-xs text-amber-600">
        {{ channelHint(channel) }}
      </div>
    </button>
  </div>
  <div v-else-if="props.showBalanceOption" class="text-sm theme-text-muted">
    {{ t('payment.channelEmptyUseBalance') }}
  </div>
  <div v-else class="text-sm theme-text-muted">
    {{ t('payment.channelEmpty') }}
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { getImageUrl } from '../../utils/image'

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const { t } = useI18n()

const props = defineProps<{
  channels: any[]
  modelValue: number | null
  showBalanceOption: boolean
  formatChannelFeeRate: (channel?: any) => string
  formatChannelFixedFee: (channel?: any) => string
  isChannelDisabledForAmount?: (channel?: any) => boolean
  channelAmountLimitHint?: (channel?: any) => string
}>()

const isDisabled = (channel?: any) => {
  if (!props.isChannelDisabledForAmount) return false
  return Boolean(props.isChannelDisabledForAmount(channel))
}

const channelHint = (channel?: any) => {
  if (!props.channelAmountLimitHint) return ''
  return String(props.channelAmountLimitHint(channel) || '')
}

const normalizedChannelType = (channel?: any) => String(channel?.channel_type || channel?.type || '').toLowerCase()

const normalizedChannelName = (channel?: any) => String(channel?.name || '').toLowerCase()

const channelBrand = (channel?: any) => {
  const type = normalizedChannelType(channel)
  const name = normalizedChannelName(channel)
  if (type === 'alipay' || name.includes('支付宝') || name.includes('alipay')) return 'alipay'
  if (type === 'wechat' || type === 'wxpay' || name.includes('微信') || name.includes('wechat')) return 'wechat'
  if (type === 'balance' || name.includes('余额')) return 'balance'
  return 'default'
}

const channelIconText = (channel?: any) => {
  const brand = channelBrand(channel)
  if (brand === 'alipay') return '支'
  if (brand === 'wechat') return '✓'
  if (brand === 'balance') return '¥'
  return '付'
}

const channelIconClass = (channel?: any) => {
  const brand = channelBrand(channel)
  if (brand === 'alipay') return 'bg-[#1677ff]'
  if (brand === 'wechat') return 'bg-[#22ac38]'
  if (brand === 'balance') return 'bg-[#f0b45f]'
  return 'bg-slate-500'
}

const selectedDotClass = (channel?: any) => {
  const brand = channelBrand(channel)
  if (brand === 'alipay') return 'bg-[#1677ff]'
  if (brand === 'wechat') return 'bg-[#22ac38]'
  if (brand === 'balance') return 'bg-[#f0b45f]'
  return 'bg-primary'
}

const channelButtonClass = (channel?: any) => {
  const selected = props.modelValue === channel?.id && !isDisabled(channel)
  const brand = channelBrand(channel)
  if (!selected) return 'theme-interactive-surface hover:border-[#1677ff]/40'
  if (brand === 'alipay') return 'border-[#1677ff] bg-[#1677ff]/10'
  if (brand === 'wechat') return 'border-[#22ac38] bg-[#22ac38]/10'
  if (brand === 'balance') return 'border-[#f0b45f] bg-[#f0b45f]/10'
  return 'theme-selected-surface'
}

const channelTextClass = (channel?: any) => {
  const selected = props.modelValue === channel?.id && !isDisabled(channel)
  const brand = channelBrand(channel)
  if (selected && brand === 'alipay') return 'text-[#1677ff]'
  if (selected && brand === 'wechat') return 'text-[#22ac38]'
  if (selected && brand === 'balance') return 'text-[#b7791f]'
  return 'theme-text-primary'
}

const channelFeeLabel = (channel?: any) => {
  const rate = props.formatChannelFeeRate(channel)
  const fixed = props.formatChannelFixedFee(channel)
  const parts = []
  if (rate && rate !== '0%' && rate !== '0.00%' && rate !== '0') parts.push(rate)
  if (fixed && !/^0(?:\.0+)?(?:\s*[A-Z]+)?$/i.test(fixed.trim())) parts.push(fixed)
  return parts.length > 0 ? `+${parts.join('+')}` : ''
}

const handleSelect = (channel?: any) => {
  if (!channel || isDisabled(channel)) return
  const id = Number(channel.id)
  if (!Number.isFinite(id) || id <= 0) return
  emit('update:modelValue', id)
}
</script>
