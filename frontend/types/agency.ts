export interface Agency {
  id: number
  name: string
  description: string
  department: string
  city: string
  address: string
  phone: string
  email: string
  manager_id: number
  manager_name: string
  manager_email: string
  manager_status: string
  status: 'active' | 'inactive'
  published: boolean
  minimum_paying_age: number
  accepts_qr: boolean
  accepts_transfer: boolean
  bank_name: string
  account_holder: string
  account_number: string
  payment_instructions: string
  qr_image?: string
  version: number
  created_at: string
  updated_at: string
}
export interface AgencyManager { id: number, first_name: string, last_name: string, email: string }
