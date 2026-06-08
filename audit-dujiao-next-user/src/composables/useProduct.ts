import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../stores/app'
import { amountToCents, centsToAmount } from '../utils/money'

export function useLocalized() {
  const appStore = useAppStore()

  const getLocalizedText = (jsonData: any): string => {
    if (!jsonData) return ''
    const locale = appStore.locale
    return jsonData[locale] || jsonData['zh-CN'] || jsonData['en-US'] || ''
  }

  const siteCurrency = computed(() => {
    const raw = String(appStore.config?.currency || '').trim().toUpperCase()
    return /^[A-Z]{3}$/.test(raw) ? raw : 'CNY'
  })

  const formatPrice = (amount: any, currency?: any): string => {
    const cur = currency ?? siteCurrency.value
    if (amount === null || amount === undefined || amount === '') return '-'
    if (cur === null || cur === undefined || cur === '') {
      return String(amount)
    }
    return `${amount} ${cur}`
  }

  return { getLocalizedText, siteCurrency, formatPrice }
}

export function useProductLabels() {
  const { t } = useI18n()

  const getPurchaseTypeLabel = (purchaseType: string) => {
    return purchaseType === 'guest' ? t('productPurchase.guest') : t('productPurchase.member')
  }

  const getFulfillmentTypeLabel = (fulfillmentType: string) => {
    return fulfillmentType === 'auto' ? t('products.fulfillmentType.auto') : t('products.fulfillmentType.manual')
  }

  const getStockBadgeClass = (status: string) => {
    switch (status) {
      case 'unlimited':
        return 'theme-badge-info'
      case 'low_stock':
        return 'theme-badge-warning'
      case 'out_of_stock':
        return 'theme-badge-danger'
      default:
        return 'theme-badge-success'
    }
  }

  const getStockStatusLabel = (product: any) => {
    const status = product?.stock_status || ''
    if (status === 'unlimited') return t('products.stockStatus.unlimited')
    if (status === 'out_of_stock') return t('products.stockStatus.outOfStock')
    if (status === 'low_stock') {
      const count = Number(product?.fulfillment_type === 'manual' ? product?.manual_stock_available : product?.auto_stock_available)
      if (Number.isFinite(count) && count > 0) {
        return t('products.stockStatus.lowStockCount', { count })
      }
      return t('products.stockStatus.lowStock')
    }
    return t('products.stockStatus.inStock')
  }

  const isSoldOut = (product: any) => Boolean(product?.is_sold_out || product?.stock_status === 'out_of_stock')

  const parsePriceAmount = (amount: any) => amountToCents(amount)

  const getPromotionPriceAmount = (product: any) => product?.promotion_price_amount

  const hasPromotionPrice = (product: any) => {
    if (!product) return false
    const original = parsePriceAmount(product.price_amount)
    const promotion = parsePriceAmount(product.promotion_price_amount)
    if (original === null || promotion === null) return false
    return promotion >= 0 && promotion < original
  }

  const getPromotionSaveAmount = (product: any) => {
    const original = parsePriceAmount(product?.price_amount)
    const promotion = parsePriceAmount(product?.promotion_price_amount)
    if (original === null || promotion === null || promotion >= original) {
      return '0.00'
    }
    return centsToAmount(original - promotion)
  }

  const hasSkuPromotionPrice = (sku: any) => {
    if (!sku) return false
    const original = parsePriceAmount(sku.price_amount)
    const promotion = parsePriceAmount(sku.promotion_price_amount)
    if (original === null || promotion === null) return false
    return promotion >= 0 && promotion < original
  }

  const getSkuPromotionPriceAmount = (sku: any) => sku?.promotion_price_amount

  const getSkuPromotionSaveAmount = (sku: any) => {
    const original = parsePriceAmount(sku?.price_amount)
    const promotion = parsePriceAmount(sku?.promotion_price_amount)
    if (original === null || promotion === null || promotion >= original) {
      return '0.00'
    }
    return centsToAmount(original - promotion)
  }

  const hasPromotionRules = (product: any) => product?.promotion_rules?.length > 0
  const getPromotionRules = (product: any): any[] => product?.promotion_rules ?? []
  const getWholesalePrices = (product: any): any[] => {
    if (Array.isArray(product?.wholesale_prices)) return product.wholesale_prices
    if (Array.isArray(product?.wholesalePrices)) return product.wholesalePrices
    return []
  }
  const hasWholesalePrices = (product: any) => getWholesalePrices(product).length > 0

  const resolveWholesalePriceAmount = (product: any, basePrice: any, quantity: number) => {
    const base = parsePriceAmount(basePrice)
    if (base === null || !Number.isFinite(quantity) || quantity <= 0) return null
    let matched: any = null
    for (const tier of getWholesalePrices(product)) {
      const minQuantity = Number(tier?.min_quantity || 0)
      const priceCents = parsePriceAmount(tier?.unit_price)
      if (!Number.isFinite(minQuantity) || minQuantity <= 0 || priceCents === null) continue
      if (quantity >= minQuantity && (!matched || minQuantity > Number(matched.min_quantity || 0))) {
        matched = tier
      }
    }
    if (!matched) return null
    const tierCents = parsePriceAmount(matched.unit_price)
    if (tierCents === null || tierCents >= base) return null
    return centsToAmount(tierCents)
  }

  const resolveMemberPriceAmount = (product: any, skuId: any, basePrice: any, memberLevelId: any, discountRate?: any) => {
    const base = parsePriceAmount(basePrice)
    const levelId = Number(memberLevelId || 0)
    if (base === null || base <= 0 || !Number.isFinite(levelId) || levelId <= 0) return null

    const prices = Array.isArray(product?.member_prices) ? product.member_prices : []
    const normalizedSkuId = Number(skuId || 0)
    const skuPrice = prices.find((p: any) => Number(p?.member_level_id || 0) === levelId && Number(p?.sku_id || 0) === normalizedSkuId)
    const productPrice = prices.find((p: any) => Number(p?.member_level_id || 0) === levelId && Number(p?.sku_id || 0) === 0)
    const overrideCents = parsePriceAmount(skuPrice?.price_amount ?? productPrice?.price_amount)
    if (overrideCents !== null && overrideCents > 0) {
      return overrideCents < base ? centsToAmount(overrideCents) : null
    }

    const rate = Number(discountRate || 0)
    if (!Number.isFinite(rate) || rate <= 0 || rate >= 100) return null
    const memberCents = Math.round((base * rate) / 100)
    return memberCents < base ? centsToAmount(memberCents) : null
  }

  return {
    getPurchaseTypeLabel,
    getFulfillmentTypeLabel,
    getStockBadgeClass,
    getStockStatusLabel,
    isSoldOut,
    hasPromotionPrice,
    getPromotionPriceAmount,
    getPromotionSaveAmount,
    hasSkuPromotionPrice,
    getSkuPromotionPriceAmount,
    getSkuPromotionSaveAmount,
    hasPromotionRules,
    getPromotionRules,
    hasWholesalePrices,
    getWholesalePrices,
    resolveWholesalePriceAmount,
    resolveMemberPriceAmount,
  }
}
