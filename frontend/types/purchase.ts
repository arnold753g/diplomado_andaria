export type PurchaseStatus = 'payment_review' | 'correction_requested' | 'confirmed' | 'payment_rejected' | 'cancelled' | 'refund_pending' | 'refunded'
export type PaymentMethod = 'qr' | 'transfer'

export interface PurchaseMinor {
  age: number
  is_foreign: boolean
  pays: boolean
}

export interface Purchase {
  id: number
  reference: string
  tourist_id: number
  agency_id: number
  package_id: number
  departure_id: number
  status: PurchaseStatus
  payment_method: PaymentMethod
  national_adults: number
  foreign_adults: number
  minors: PurchaseMinor[]
  free_minor_count: number
  paying_minor_count: number
  capacity_count: number
  national_unit_price_cents: number
  foreign_surcharge_cents: number
  total_cents: number
  tourist_name: string
  tourist_email: string
  tourist_phone: string
  tourist_document: string
  tourist_nationality: string
  agency_review_user_id?: number
  reviewed_at?: string
  rejection_reason: string
  cancelled_at?: string
  cancellation_reason: string
  refund_reason: string
  refund_requested_at?: string
  refund_due_at?: string
  refund_method?: 'qr' | 'bank_transfer'
  refund_bank_name?: string
  refund_account_holder?: string
  refund_account_number?: string
  refund_reference?: string
  refunded_at?: string
  refund_completed_by_user_id?: number
  version: number
  package_name: string
  agency_name: string
  departure_start: string
  meeting_point: string
  has_proof: boolean
  has_refund_qr: boolean
  has_refund_proof: boolean
  cancellable: boolean
  cancellation_deadline?: string
  refund_overdue: boolean
  created_at: string
  updated_at: string
}

export interface PurchasePage {
  purchases: Purchase[]
  pagination: { page: number; limit: number; total: number }
}

export interface PurchasePaymentOptions {
  package_id: number
  agency_name: string
  minimum_paying_age: number
  methods: PaymentMethod[]
  bank_name: string
  account_holder: string
  account_number: string
  payment_instructions: string
  qr_image?: string
}
