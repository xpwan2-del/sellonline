const PRODUCTION_API_BASE_URL = 'https://api.toplenged.com'
const PRODUCTION_HOSTS = new Set(['toplenged.com', 'www.toplenged.com'])

export function getApiBaseUrl(): string {
    if (typeof window !== 'undefined' && PRODUCTION_HOSTS.has(window.location.hostname)) {
        return PRODUCTION_API_BASE_URL
    }
    return import.meta.env.VITE_API_BASE_URL || ''
}
