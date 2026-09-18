export interface PackagePhoto { id: number; position: number }
export interface PackageAttractionSummary { id: number; name: string; department: string; city: string }
export interface PackageItineraryAttraction { attraction_id: number; position: number; attraction?: PackageAttractionSummary }
export interface PackageItineraryDay {
  id: number
  day_number: number
  title: string
  description: string
  activities: string[]
  attractions: PackageItineraryAttraction[]
}
export type PackageFrequency = 'single' | 'daily' | 'specific_weekdays'
export interface PackageSchedule {
  id: number
  frequency_type: PackageFrequency
  valid_from: string
  valid_until: string
  weekdays: number[]
  departure_time: string
  meeting_time: string
  default_min_capacity: number
  default_max_capacity: number
  booking_cutoff_hours: number
  maximum_advance_days: number
  default_meeting_point: string
  default_instructions: string
}
export interface PackageDeparture {
  id: number
  package_id: number
  schedule_id?: number
  starts_at: string
  meeting_at: string
  booking_opens_at: string
  booking_closes_at: string
  min_capacity: number
  max_capacity: number
  held_capacity: number
  confirmed_capacity: number
  meeting_point: string
  instructions: string
  status: 'draft' | 'open' | 'confirmed' | 'closed' | 'minimum_review' | 'cancelled' | 'completed'
  cancellation_reason: string
  is_exception: boolean
  modified_by_user_id?: number
  cancelled_at?: string
  minimum_review_at?: string
  minimum_reviewed_at?: string
  minimum_review_user_id?: number
  version: number
}
export interface PublicPackageDeparture {
  id: number
  starts_at: string
  meeting_at: string
  booking_opens_at: string
  booking_closes_at: string
  min_capacity: number
  max_capacity: number
  confirmed_capacity: number
  available_capacity: number
  remaining_for_minimum: number
  minimum_reached: boolean
  meeting_point: string
  instructions: string
  status: 'draft' | 'open' | 'confirmed' | 'closed'
  bookable: boolean
}
export interface PackageCatalogItem {
  id: number
  agency_id: number
  agency_name: string
  agency_department: string
  agency_city: string
  name: string
  description: string
  duration_days: number
  duration_nights: number
  difficulty: '' | 'easy' | 'moderate' | 'demanding'
  national_price_cents: number
  foreign_surcharge_cents: number
  photos: PackagePhoto[]
  frequency_type: PackageFrequency | ''
  next_departure?: PublicPackageDeparture
  bookable: boolean
}
export interface PublicPackageDetail extends PackageCatalogItem {
  includes: string[]
  excludes: string[]
  bring: string[]
  cancellation_allowed: boolean
  cancellation_notice_hours: number
  minimum_paying_age: number
  itinerary: PackageItineraryDay[]
  departures: PublicPackageDeparture[]
}
export interface PackageCatalogPage { packages: PackageCatalogItem[]; pagination: { page: number; limit: number; total: number } }
export interface PackageCatalogOptions { departments: string[]; difficulties: Array<{ label: string; value: string }> }
export interface TourPackage {
  id: number
  agency_id: number
  agency_name: string
  minimum_paying_age: number
  name: string
  description: string
  duration_days: number
  duration_nights: number
  difficulty: '' | 'easy' | 'moderate' | 'demanding'
  national_price_cents: number
  foreign_surcharge_cents: number
  includes: string[]
  excludes: string[]
  bring: string[]
  cancellation_allowed: boolean
  cancellation_notice_hours: number
  published: boolean
  version: number
  photos: PackagePhoto[]
  itinerary: PackageItineraryDay[]
  schedule?: PackageSchedule
  departures?: PackageDeparture[]
  created_at: string
  updated_at: string
}
export interface PackagePage { packages: TourPackage[]; pagination: { page: number; limit: number; total: number } }
