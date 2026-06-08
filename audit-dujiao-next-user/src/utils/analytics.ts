import type { RouteLocationNormalizedLoaded, RouteLocationNormalized } from 'vue-router'
import { analyticsAPI } from '../api/analytics'
import { getAffiliateCode, getAffiliateVisitorKey } from './affiliate'

const VISIT_SESSION_KEY = 'dj_visit_session_key'

const ensureSessionKey = () => {
    if (typeof window === 'undefined') return ''
    const existing = String(sessionStorage.getItem(VISIT_SESSION_KEY) || '').trim()
    if (existing) return existing
    const next = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 12)}`
    sessionStorage.setItem(VISIT_SESSION_KEY, next)
    return next
}

const getCurrentUserID = () => {
    if (typeof window === 'undefined') return 0
    const raw = localStorage.getItem('user_profile')
    if (!raw) return 0
    try {
        const parsed = JSON.parse(raw)
        const id = Number(parsed?.id || 0)
        return Number.isFinite(id) && id > 0 ? Math.floor(id) : 0
    } catch {
        return 0
    }
}

const resolveSourceType = (affiliateCode: string) => {
    if (affiliateCode) return 'affiliate'
    if (typeof document === 'undefined' || !document.referrer) return 'direct'
    try {
        const referrer = new URL(document.referrer)
        if (referrer.host === window.location.host) return 'direct'
        const host = referrer.hostname.toLowerCase()
        if (
            host.includes('google.') ||
            host.includes('bing.') ||
            host.includes('baidu.') ||
            host.includes('yahoo.') ||
            host.includes('duckduckgo.') ||
            host.includes('sogou.') ||
            host.includes('so.com') ||
            host.includes('yandex.')
        ) {
            return 'search'
        }
        return 'external'
    } catch {
        return 'unknown'
    }
}

const resolveDeviceType = () => {
    if (typeof navigator === 'undefined') return 'unknown'
    const ua = navigator.userAgent.toLowerCase()
    if (ua.includes('ipad') || ua.includes('tablet')) return 'tablet'
    if (ua.includes('mobile') || ua.includes('iphone') || ua.includes('android')) return 'mobile'
    return 'desktop'
}

const resolvePageType = (path: string) => {
    if (path === '/') return 'home'
    if (path.startsWith('/products/')) return 'product'
    if (path.startsWith('/categories/')) return 'category'
    if (path.startsWith('/cart')) return 'cart'
    if (path.startsWith('/checkout') || path.startsWith('/pay')) return 'checkout'
    if (path.startsWith('/me')) return 'account'
    return 'page'
}

export const trackSiteVisit = async (route: RouteLocationNormalizedLoaded | RouteLocationNormalized) => {
    if (typeof window === 'undefined') return

    const affiliateCode = getAffiliateCode()
    try {
        await analyticsAPI.trackVisit({
            visitor_key: getAffiliateVisitorKey(),
            session_key: ensureSessionKey(),
            user_id: getCurrentUserID(),
            path: route.fullPath || route.path || '/',
            page_type: resolvePageType(route.path || '/'),
            source_type: resolveSourceType(affiliateCode),
            referrer: document.referrer || '',
            affiliate_code: affiliateCode,
            device_type: resolveDeviceType(),
        })
    } catch {
    }
}
