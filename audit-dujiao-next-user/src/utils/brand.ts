export const DEFAULT_SITE_NAME = 'BuyBuyGPT'

const legacyBrandPattern = /\btop[\s_-]*(?:legend|lenged)\b/gi

export const normalizeBrandText = (value: unknown): string => {
  return String(value || '').replace(legacyBrandPattern, DEFAULT_SITE_NAME).trim()
}

export const resolveSiteName = (value: unknown): string => {
  return normalizeBrandText(value) || DEFAULT_SITE_NAME
}
