export type Role = 'admin' | 'turista' | 'encargado_agencia' | 'encargado_atraccion'
export type UserStatus = 'active' | 'inactive'

export interface User {
  id: number
  email: string
  first_name: string
  last_name: string
  phone: string
  document_number: string
  nationality: string
  has_password: boolean
  google_linked: boolean
  role: Role
  status: UserStatus
  last_login_at: string | null
  created_at: string
  updated_at: string
}

export interface SessionData {
  user: User
  csrf_token: string
  session_expires_at: string
  idle_timeout_seconds: number
}
