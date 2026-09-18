import assert from 'node:assert/strict'
import test from 'node:test'
import type { PublicPackageDeparture } from '../types/package.ts'
import { packageDepartureStatus, packageDuration, packageMoney } from '../utils/packages.ts'

const departure = (overrides: Partial<PublicPackageDeparture> = {}): PublicPackageDeparture => ({
  id: 1,
  starts_at: '2026-10-10T12:00:00Z',
  meeting_at: '2026-10-10T11:30:00Z',
  booking_opens_at: '2026-09-01T00:00:00Z',
  booking_closes_at: '2026-10-09T12:00:00Z',
  min_capacity: 4,
  max_capacity: 12,
  confirmed_capacity: 2,
  available_capacity: 10,
  remaining_for_minimum: 2,
  minimum_reached: false,
  meeting_point: '',
  instructions: '',
  status: 'open',
  bookable: true,
  ...overrides
})

test('package catalog formats commercial values consistently', () => {
  assert.match(packageMoney(25000), /250[,.]00/)
  assert.equal(packageDuration({ duration_days: 2, duration_nights: 1 }), '2 días · 1 noche')
})

test('departure status prioritizes sold out and minimum state', () => {
  assert.equal(packageDepartureStatus(departure({ available_capacity: 0 })), 'Agotada')
  assert.equal(packageDepartureStatus(departure()), 'Venta abierta')
  assert.equal(packageDepartureStatus(departure({ minimum_reached: true })), 'Confirmada y en venta')
})
