import assert from 'node:assert/strict'
import test from 'node:test'
import { previewPackageSchedule } from '../utils/package-schedule.ts'

const dateKeys = (dates: Date[]) => dates.map(date => `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`)

test('single departure selected at midnight is included, regardless of the hidden end date', () => {
  const start = new Date(2026, 9, 30)
  assert.deepEqual(dateKeys(previewPackageSchedule('single', start, new Date(2026, 10, 15))), ['2026-10-30'])
})

test('daily departures include the final calendar day selected at midnight', () => {
  assert.deepEqual(dateKeys(previewPackageSchedule('daily', new Date(2026, 9, 29), new Date(2026, 9, 30))), ['2026-10-29', '2026-10-30'])
})

test('a matching weekday on a one-day range generates one departure', () => {
  assert.deepEqual(dateKeys(previewPackageSchedule('specific_weekdays', new Date(2026, 9, 30), new Date(2026, 9, 30), [5])), ['2026-10-30'])
})

test('loaded noon and selected midnight represent the same calendar day', () => {
  assert.deepEqual(dateKeys(previewPackageSchedule('daily', new Date(2026, 9, 30, 12), new Date(2026, 9, 30))), ['2026-10-30'])
})

test('weekday selection crosses month and year boundaries inclusively', () => {
  assert.deepEqual(dateKeys(previewPackageSchedule('specific_weekdays', new Date(2026, 11, 31), new Date(2027, 0, 3), [4, 7])), ['2026-12-31', '2027-1-3'])
})

test('missing, reversed and invalid dates do not generate departures', () => {
  assert.deepEqual(previewPackageSchedule('single', null, null), [])
  assert.deepEqual(previewPackageSchedule('daily', new Date(2026, 9, 30), null), [])
  assert.deepEqual(previewPackageSchedule('daily', new Date(2026, 10, 1), new Date(2026, 9, 30)), [])
  assert.deepEqual(previewPackageSchedule('single', new Date('invalid'), null), [])
  assert.deepEqual(previewPackageSchedule('specific_weekdays', new Date(2026, 9, 30), new Date(2026, 9, 30), [1]), [])
})
