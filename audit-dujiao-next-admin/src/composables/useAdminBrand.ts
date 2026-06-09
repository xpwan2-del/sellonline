import { computed, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { applySiteIcon } from '@/utils/favicon'

interface AdminBrandPublicConfig {
  app_version?: string
  brand?: {
    site_name?: unknown
    site_url?: unknown
    site_icon?: unknown
  }
  captcha?: unknown
}

const envBrandName = String(import.meta.env.VITE_ADMIN_BRAND_NAME || '').trim()
const envSiteURL = normalizeURL(import.meta.env.VITE_ADMIN_SITE_URL || '')

const siteName = ref(envBrandName)
const siteURL = ref(envSiteURL)
const appVersion = ref('')
let publicConfigPromise: Promise<AdminBrandPublicConfig | null> | null = null

function normalizeText(value: unknown): string {
  return String(value || '').trim()
}

function normalizeURL(value: unknown): string {
  return normalizeText(value).replace(/\/+$/, '')
}

function resolveBrandName(value: unknown): string {
  const name = normalizeText(value)
  return name || envBrandName || 'Admin'
}

function applyPublicConfig(payload: AdminBrandPublicConfig | null | undefined) {
  const brand = payload?.brand || {}
  siteName.value = resolveBrandName(brand.site_name)
  siteURL.value = normalizeURL(brand.site_url) || envSiteURL
  applySiteIcon(brand.site_icon)

  if (typeof payload?.app_version === 'string') {
    appVersion.value = payload.app_version
  }

  document.title = adminBrandName.value
}

async function loadPublicConfig(): Promise<AdminBrandPublicConfig | null> {
  if (!publicConfigPromise) {
    publicConfigPromise = adminAPI.getPublicConfig()
      .then((res) => {
        const payload = (res.data?.data || null) as AdminBrandPublicConfig | null
        applyPublicConfig(payload)
        return payload
      })
      .catch(() => {
        applyPublicConfig(null)
        return null
      })
      .finally(() => {
        publicConfigPromise = null
      })
  }

  return publicConfigPromise
}

const adminBrandName = computed(() => {
  const name = siteName.value || 'Admin'
  return /\badmin\b/i.test(name) ? name : `${name} Admin`
})

export function useAdminBrand() {
  return {
    appVersion,
    siteName,
    siteURL,
    adminBrandName,
    loadPublicConfig,
  }
}
