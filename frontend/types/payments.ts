import type { PackageDeparture } from './package'
export type AuthorizationState = 'authorized' | 'ready' | 'minimum_pending' | 'minimum_review' | 'cancelled' | 'past' | 'unavailable'
export interface PaymentDeparture extends PackageDeparture {
  package_name: string
  authorized_by_name: string
  agency_active: boolean
  authorization_state: AuthorizationState
  can_authorize: boolean
  can_cancel: boolean
  total_purchases: number
  payments_pending: number
  corrections_pending: number
  confirmed_purchases: number
  refunds_pending: number
  refunds_overdue: number
  refunded_purchases: number
  confirmed_amount_cents: number
  refund_amount_cents: number
  confirmed_travelers: number
  confirmed_adults: number
  confirmed_minors: number
  confirmed_free_minors: number
}
export interface PaymentsOverview {
  departures: PaymentDeparture[]
  packages: { id: number; name: string }[]
  totals: { departures: number; payments_pending: number; corrections_pending: number; refunds_pending: number; refunds_overdue: number; ready_departures: number; confirmed_amount_cents: number }
  summary: { payments_pending: number; corrections_pending: number; refunds_pending: number; refunds_overdue: number; ready_departures: number; authorized_departures: number; pending_departures: number; refund_departures: number; history_departures: number }
  pagination: { page: number; limit: number; total: number }
}
