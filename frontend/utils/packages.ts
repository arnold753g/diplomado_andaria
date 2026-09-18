import type { PackageCatalogItem, PackageFrequency, PublicPackageDeparture } from '../types/package.ts'

export const packageMoney = (cents: number) => new Intl.NumberFormat('es-BO', {
  style: 'currency', currency: 'BOB', minimumFractionDigits: 2
}).format(cents / 100)

export const packageDuration = (item: Pick<PackageCatalogItem, 'duration_days' | 'duration_nights'>) => {
  const days = `${item.duration_days} ${item.duration_days === 1 ? 'día' : 'días'}`
  const nights = item.duration_nights ? ` · ${item.duration_nights} ${item.duration_nights === 1 ? 'noche' : 'noches'}` : ''
  return days + nights
}

export const packageDifficulty = (difficulty: string) => ({
  easy: 'Fácil', moderate: 'Moderada', demanding: 'Exigente'
}[difficulty] || 'No especificada')

export const packageFrequency = (frequency: PackageFrequency | '') => frequency ? ({
  single: 'Salida única', daily: 'Salidas diarias', specific_weekdays: 'Días específicos'
}[frequency]) : 'Programación por confirmar'

export const packageDepartureDate = (value: string) => new Intl.DateTimeFormat('es-BO', {
  weekday: 'long', day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit',
  timeZone: 'America/La_Paz'
}).format(new Date(value))

export const packageDepartureShortDate = (value: string) => new Intl.DateTimeFormat('es-BO', {
  day: '2-digit', month: 'short', year: 'numeric', timeZone: 'America/La_Paz'
}).format(new Date(value))

export const packageDepartureStatus = (departure: PublicPackageDeparture) => {
  if (departure.available_capacity <= 0) return 'Agotada'
  if (departure.bookable) return departure.minimum_reached ? 'Confirmada y en venta' : 'Venta abierta'
  if (new Date(departure.booking_opens_at).getTime() > Date.now()) return 'Venta próxima'
  return 'Venta cerrada'
}
