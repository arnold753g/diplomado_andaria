import type { PackageFrequency } from '../types/package.ts'

export function previewPackageSchedule(frequency: PackageFrequency, start: Date | null, end: Date | null, weekdays: readonly number[] = []): Date[] {
  // DatePicker selections use midnight; loaded API dates may use noon.
  // Compare calendar days with the same clock so the last day is inclusive.
  const calendarDay = (value: Date | null) => value && Number.isFinite(value.getTime())
    ? new Date(value.getFullYear(), value.getMonth(), value.getDate(), 12)
    : null
  const from = calendarDay(start)
  const until = frequency === 'single' ? from : calendarDay(end)
  if (!from || !until || until < from) return []
  const result: Date[] = []
  for (let day = from; day <= until && result.length <= 366; day = new Date(day.getFullYear(), day.getMonth(), day.getDate() + 1, 12)) {
    const isoDay = day.getDay() || 7
    if (frequency === 'single' || frequency === 'daily' || weekdays.includes(isoDay)) result.push(day)
  }
  return result
}
