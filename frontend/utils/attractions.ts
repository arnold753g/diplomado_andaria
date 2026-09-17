import type { Attraction } from "../types/attraction.ts"
export const admissionLabel = (cents: number) => cents === 0 ? 'Entrada gratuita' : `Bs ${(cents / 100).toFixed(2)}`

export const googleMapsDirectionsURL = (latitude: number, longitude: number) => {
  const destination = `${latitude},${longitude}`
  return `https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(destination)}&travelmode=driving`
}

export const weekDays = ['Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado', 'Domingo'].map((name, i) => ({ id: i + 1, name }))
export const months = ['Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio', 'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'].map((name, i) => ({ id: i + 1, name }))
export const scheduleLabel = (a: Pick<Attraction, 'schedule_mode' | 'opening_time' | 'closing_time' | 'opening_days' | 'opening_hours'>) => {
  if (a.schedule_mode === 'unspecified') return a.opening_hours || 'Consulta los horarios antes de visitar.'
  const days = a.opening_days.length === 7 ? 'Todos los días' : [...a.opening_days].sort((x, y) => x - y).map(id => weekDays.find(d => d.id === id)?.name).filter(Boolean).join(', ')
  if (a.schedule_mode === 'all_day') return `${days} · 24 horas`
  return `${days} · ${a.opening_time} a ${a.closing_time}${a.closing_time < a.opening_time ? ' (cierre al día siguiente)' : ''}`
}
export const seasonLabel = (a: Pick<Attraction, 'season_mode' | 'season_start_month' | 'season_end_month'>) => {
  if (a.season_mode === 'all_year') return 'Todo el año'
  if (a.season_mode !== 'months') return 'Consulta la mejor época antes de viajar.'
  const start = months.find(m => m.id === a.season_start_month)?.name || ''
  const end = months.find(m => m.id === a.season_end_month)?.name || ''
  if (a.season_start_month === a.season_end_month) return start
  return `${start} a ${end}${(a.season_end_month || 0) < (a.season_start_month || 0) ? ' (hasta el año siguiente)' : ''}`
}
