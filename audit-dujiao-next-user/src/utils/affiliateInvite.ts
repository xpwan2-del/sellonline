import type { RouteLocationNormalized, RouteLocationNormalizedLoaded } from 'vue-router'
import { affiliateAPI } from '../api/affiliate'

const AGENT_INVITE_STORAGE_KEY = 'dj_agent_invite'

const normalizeAgentInviteCode = (raw: string) => {
    const value = String(raw || '').trim().toUpperCase()
    if (!/^[A-Z0-9]{4,32}$/.test(value)) {
        return ''
    }
    return value
}

const resolveQueryCode = (route: RouteLocationNormalizedLoaded | RouteLocationNormalized) => {
    const raw = Array.isArray(route.query.agent_invite) ? route.query.agent_invite[0] : route.query.agent_invite
    if (!raw) return ''
    return normalizeAgentInviteCode(String(raw))
}

export const captureAgentInviteFromRoute = (route: RouteLocationNormalizedLoaded | RouteLocationNormalized) => {
    if (typeof window === 'undefined') return ''
    const code = resolveQueryCode(route)
    if (!code) return getAgentInviteCode()
    localStorage.setItem(AGENT_INVITE_STORAGE_KEY, code)
    return code
}

export const getAgentInviteCode = () => {
    if (typeof window === 'undefined') return ''
    return normalizeAgentInviteCode(localStorage.getItem(AGENT_INVITE_STORAGE_KEY) || '')
}

export const clearAgentInviteCode = () => {
    if (typeof window === 'undefined') return
    localStorage.removeItem(AGENT_INVITE_STORAGE_KEY)
}

export const checkAgentInviteCode = async (code: string) => {
    const normalized = normalizeAgentInviteCode(code)
    if (!normalized) return false
    const response = await affiliateAPI.checkInviteCode(normalized)
    return response.data?.data?.valid === true
}
