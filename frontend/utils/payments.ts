import type { AuthorizationState, PaymentDeparture } from '../types/payments.ts'
import type { Purchase } from '../types/purchase.ts'
export const authorizationLabel = (state: AuthorizationState) => ({
  authorized: 'Salida autorizada', ready: 'Lista para confirmar', minimum_pending: 'Mínimo pendiente',
  minimum_review: 'Decisión por mínimo', cancelled: 'Cancelada', past: 'Salida pasada', unavailable: 'No habilitada'
}[state])
export const authorizationSeverity = (state: AuthorizationState): 'success' | 'warn' | 'danger' | 'secondary' | 'info' => ({
  authorized: 'success', ready: 'info', minimum_pending: 'secondary', minimum_review: 'warn', cancelled: 'danger', past: 'secondary', unavailable: 'secondary'
}[state] as 'success' | 'warn' | 'danger' | 'secondary' | 'info')
export const purchaseTravelers = (purchase: Pick<Purchase, 'national_adults' | 'foreign_adults' | 'minors'>) => purchase.national_adults + purchase.foreign_adults + purchase.minors.length
export const groupPaymentDepartures = (rows: PaymentDeparture[]) => {
  const groups = new Map<number, { id: number; name: string; departures: PaymentDeparture[] }>()
  for (const row of rows) {
    if (!groups.has(row.package_id)) groups.set(row.package_id, { id: row.package_id, name: row.package_name, departures: [] })
    groups.get(row.package_id)!.departures.push(row)
  }
  return [...groups.values()]
}
export const paymentDateTime = (value: string) => new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short', hourCycle: 'h23', timeZone: 'America/La_Paz' }).format(new Date(value))
export type PaymentView = 'authorized' | 'pending' | 'refunds' | 'history'
export const paymentViews: { value: PaymentView; label: string; icon: string; counter: 'authorized_departures' | 'pending_departures' | 'refund_departures' | 'history_departures' }[] = [
  { value: 'authorized', label: 'Habilitadas', icon: 'pi pi-check-circle', counter: 'authorized_departures' },
  { value: 'pending', label: 'Pendientes', icon: 'pi pi-clock', counter: 'pending_departures' },
  { value: 'refunds', label: 'Reembolsos', icon: 'pi pi-wallet', counter: 'refund_departures' },
  { value: 'history', label: 'Historial', icon: 'pi pi-history', counter: 'history_departures' }
]
export const normalizePaymentView = (value: unknown): PaymentView => typeof value === 'string' && paymentViews.some(item => item.value === value) ? value as PaymentView : 'authorized'
export type CancellationCause = 'minimum' | 'external'
export const departureCancellationReason = (cause: CancellationCause, detail: string) => `${cause === 'minimum' ? 'Falta de cupos' : 'Motivos externos'}: ${detail.trim()}`
