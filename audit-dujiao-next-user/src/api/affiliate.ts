import { api, userApi } from './client'
import type { AffiliateAgentApplicationPayload, AffiliateWithdrawApplyPayload } from './types'

export const affiliateAPI = {
    trackClick: (data: { affiliate_code: string; visitor_key?: string; landing_path?: string; referrer?: string }) =>
        api.post('/public/affiliate/click', data),
    checkInviteCode: (code: string) => api.get(`/public/affiliate-invite-codes/${encodeURIComponent(code)}/check`),
    open: () => userApi.post('/affiliate/open'),
    dashboard: () => userApi.get('/affiliate/dashboard'),
    agentApplication: () => userApi.get('/affiliate/application'),
    applyAgent: (data: AffiliateAgentApplicationPayload) => userApi.post('/affiliate/applications', data),
    reportSummary: (params?: any) => userApi.get('/affiliate/report/summary', { params }),
    reportCommissions: (params?: any) => userApi.get('/affiliate/report/commissions', { params }),
    customers: (params?: any) => userApi.get('/affiliate/customers', { params }),
    commissions: (params?: any) => userApi.get('/affiliate/commissions', { params }),
    withdraws: (params?: any) => userApi.get('/affiliate/withdraws', { params }),
    applyWithdraw: (data: AffiliateWithdrawApplyPayload) =>
        userApi.post('/affiliate/withdraws', data),
}
