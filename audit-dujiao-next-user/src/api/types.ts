export interface UserProfileData {
    id: number
    email: string
    nickname: string
    email_verified_at?: string | null
    locale: string
    member_level_id?: number
    total_recharged?: number | string
    total_spent?: number | string
    email_change_mode?: 'bind_only' | 'change_with_old_and_new'
    password_change_mode?: 'set_without_old' | 'change_with_old'
}

export interface PublicMemberLevel {
    id: number
    name: Record<string, string>
    slug: string
    icon: string
    discount_rate: number
    recharge_threshold: number
    spend_threshold: number
    is_default: boolean
    sort_order: number
}

export interface UpdateUserProfilePayload {
    nickname?: string
    locale?: string
}

export interface UserLoginLogItem {
    id: number
    user_id: number
    email: string
    status: string
    fail_reason?: string
    client_ip?: string
    user_agent?: string
    login_source?: string
    created_at?: string
}

export interface SendChangeEmailCodePayload {
    kind: 'old' | 'new'
    new_email?: string
}

export interface ChangeEmailPayload {
    new_email: string
    old_code?: string
    new_code: string
}

export interface ChangeUserPasswordPayload {
    old_password?: string
    new_password: string
}

export interface TelegramAuthPayload {
    id: number
    first_name?: string
    last_name?: string
    username?: string
    photo_url?: string
    auth_date: number
    hash: string
}

export interface TelegramMiniAppAuthPayload {
    init_data: string
}

export interface UserRegisterPayload {
    email: string
    password: string
    code?: string
    agreement_accepted: boolean
    agent_invite?: string
    affiliate_code?: string
    affiliate_visitor_key?: string
}

export interface TelegramBindingData {
    bound: boolean
    provider?: string
    provider_user_id?: string
    username?: string
    avatar_url?: string
    auth_at?: string | null
    updated_at?: string | null
}

export interface WalletAccountData {
    balance: string
}

export interface WalletTransactionData {
    id: number
    type: string
    direction: string
    amount: string
    balance_after: string
    remark: string
    created_at: string
}

export interface WalletRechargePayload {
    amount: string
    channel_id: number
    currency?: string
    remark?: string
}

export interface WalletRechargeOrderData {
    id: number
    recharge_no: string
    amount: string
    payable_amount: string
    fee_amount: string
    currency: string
    status: string
    remark: string
    paid_at?: string
    created_at: string
}

export interface WalletRechargeResult {
    recharge?: WalletRechargeOrderData
    recharge_no?: string
    recharge_status?: string
    account?: WalletAccountData
    payment_id?: number
    provider_type?: string
    channel_type?: string
    interaction_mode?: string
    pay_url?: string
    qr_code?: string
    expires_at?: string
    status?: string
}

export interface GiftCardData {
    id: number
    name: string
    code: string
    amount: string
    currency: string
    status: string
    redeemed_at?: string
}

export interface GiftCardRedeemResult {
    gift_card: GiftCardData
    wallet: WalletAccountData
    transaction: WalletTransactionData
    wallet_delta: string
}

export interface AffiliateDashboardData {
    opened: boolean
    affiliate_code: string
    promotion_path: string
    click_count: number
    valid_order_count: number
    conversion_rate: number
    pending_commission: string
    available_commission: string
    withdrawn_commission: string
}

export interface AffiliateAgentApplicationData {
    id: number
    user_id: number
    email: string
    phone: string
    message: string
    status: string
    admin_note?: string
    created_at: string
    updated_at: string
}

export interface AffiliateAgentApplicationPayload {
    phone: string
    message?: string
}

export interface AffiliateCommissionData {
    id: number
    commission_type: string
    order?: {
        id: number
        order_no: string
    }
    buyer_email?: string
    buyer_masked?: string
    base_amount?: string
    rate_percent?: string
    commission_amount: string
    status: string
    confirm_at?: string
    available_at?: string
    created_at: string
    commission_items?: Array<{
        id: number
        order_item_id: number
        product_id: number
        product_title?: Record<string, string>
        sku_snapshot?: { spec_values?: Record<string, unknown>; sku_code?: string; [key: string]: unknown }
        quantity: number
        base_amount: string
        rate_percent: string
        commission_amount: string
    }>
}

export interface AffiliateReportSummary {
    affiliate_code?: string
    promotion_url?: string
    total_sales_amount?: string
    commission_base_amount: string
    total_commission: string
    available_commission: string
    pending_commission: string
    withdrawn_commission: string
    rejected_commission: string
    valid_order_count: number
    customer_count?: number
    new_customer_order_count?: number
    repeat_customer_order_count?: number
    click_count: number
    conversion_rate: string
    average_commission: string
    average_commission_rate: string
}

export interface AffiliateReportTrendPoint {
    date: string
    commission_amount: string
}

export interface AffiliateReportSourceBreakdown {
    source: string
    label: string
    sales_amount: string
    commission_amount: string
}

export interface AffiliateReportCommissionItem {
    id?: number
    order_item_id?: number
    product_id: number
    product_title?: Record<string, string>
    sku_snapshot?: { spec_values?: Record<string, unknown>; sku_code?: string; [key: string]: unknown }
    quantity: number
    base_amount: string
    rate_percent: string
    commission_amount: string
    source_type: string
    source_label: string
}

export interface AffiliateReportCommissionData {
    id: number
    order_id: number
    order_no: string
    buyer_email?: string
    buyer_masked: string
    order_total_amount: string
    base_amount: string
    rate_percent: string
    commission_amount: string
    source_type: string
    source_label: string
    status: string
    created_at: string
    available_at?: string
    commission_items?: AffiliateReportCommissionItem[]
}

export interface AffiliateReportData {
    summary: AffiliateReportSummary
    trend: AffiliateReportTrendPoint[]
    source_breakdown: AffiliateReportSourceBreakdown[]
}

export interface AffiliateCustomerData {
    id: number
    email: string
    display_name: string
    registered_at: string
    last_order_at?: string
    order_count: number
    commission_amount: string
}

export interface AffiliateWithdrawData {
    id: number
    amount: string
    channel: string
    account: string
    status: string
    reject_reason?: string
    created_at: string
}

export interface AffiliateWithdrawApplyPayload {
    amount: string
    channel: string
    account: string
}

export interface CreatePaymentPayload {
    order_no: string
    channel_id?: number
    use_balance?: boolean
}

export interface PaymentCreateResult {
    order_paid?: boolean
    wallet_paid_amount?: string
    online_pay_amount?: string
    payment_id?: number
    order_no?: string
    channel_id?: number
    provider_type?: string
    channel_type?: string
    interaction_mode?: string
    pay_url?: string
    qr_code?: string
    expires_at?: string
}

export interface CaptchaPayload {
    captcha_id?: string
    captcha_code?: string
    turnstile_token?: string
}
