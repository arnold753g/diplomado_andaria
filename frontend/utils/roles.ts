import type { Role } from '../types/auth'

export const roleOptions: { label: string, value: Role }[] = [
  { label: 'Administrador', value: 'admin' },
  { label: 'Turista', value: 'turista' },
  { label: 'Encargado de agencia', value: 'encargado_agencia' },
  { label: 'Encargado de atracción', value: 'encargado_atraccion' }
]
export const roleLabel = (role?: string) => roleOptions.find(option => option.value === role)?.label || 'Cuenta'
