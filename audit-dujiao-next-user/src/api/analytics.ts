import { api } from './client'

export const analyticsAPI = {
    trackVisit: (data: any) =>
        api.post('/public/analytics/visit', data, { silentBusinessError: true }),
}
