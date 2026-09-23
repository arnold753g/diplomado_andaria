import type { PurchaseStatus } from './purchase'

export interface DashboardNextPurchase {
  reference: string
  package_name: string
  departure_start: string
  status: PurchaseStatus
}

export interface TouristDashboard {
  favorites: number
  active_purchases: number
  pending_refunds: number
  completed_refunds: number
  next_purchase?: DashboardNextPurchase
}

export interface AgencyDashboard {
  assigned: boolean
  agency_name?: string
  packages: number
  published_packages: number
  payments_to_review: number
  minimum_reviews: number
  pending_refunds: number
  refunds_due_soon: number
  overdue_refunds: number
}

export interface AttractionManagerDashboard {
  assigned_attractions: number
  published_attractions: number
  draft_attractions: number
  inactive_attractions: number
}

export interface AccountDashboard {
  role: 'turista' | 'encargado_agencia' | 'encargado_atraccion'
  tourist?: TouristDashboard
  agency?: AgencyDashboard
  attraction_manager?: AttractionManagerDashboard
}

export interface AdminDashboard {
  users: number
  active_users: number
  active_admins: number
  agencies: number
  published_agencies: number
  attractions: number
  published_attractions: number
  packages: number
  published_packages: number
  purchases: number
}
