import assert from 'node:assert/strict'
import test from 'node:test'
import { authorizationLabel, authorizationSeverity, groupPaymentDepartures, purchaseTravelers, normalizePaymentView, departureCancellationReason } from '../utils/payments.ts'
import type { PaymentDeparture } from '../types/payments.ts'

test('minimum reached and explicit authorization have distinct labels', () => {
  assert.equal(authorizationLabel('ready'), 'Lista para confirmar')
  assert.equal(authorizationLabel('authorized'), 'Salida autorizada')
  assert.notEqual(authorizationSeverity('authorized'), authorizationSeverity('minimum_review'))
})
test('traveler count includes free children independently of purchased capacity', () => {
  assert.equal(purchaseTravelers({ national_adults: 2, foreign_adults: 1, minors: [{ age: 2, is_foreign: false, pays: false }, { age: 10, is_foreign: true, pays: true }] }), 5)
})
test('departures remain separated by package identity rather than duplicated names', () => {
  const rows = [{ id: 10, package_id: 1, package_name: 'Ruta' }, { id: 11, package_id: 2, package_name: 'Ruta' }, { id: 12, package_id: 1, package_name: 'Ruta' }] as PaymentDeparture[]
  const groups = groupPaymentDepartures(rows)
  assert.deepEqual(groups.map(group => [group.id, group.departures.map(row => row.id)]), [[1, [10, 12]], [2, [11]]])
})

test('the main view defaults to future authorized departures, not the entire history', () => {
  assert.equal(normalizePaymentView(undefined), 'authorized')
  assert.equal(normalizePaymentView('all'), 'authorized')
  assert.equal(normalizePaymentView('pending'), 'pending')
  assert.equal(normalizePaymentView('refunds'), 'refunds')
  assert.equal(normalizePaymentView('history'), 'history')
})
test('the cancellation reason records the cause and trims the explanation', () => {
  assert.equal(departureCancellationReason('external', '  Corte de carretera  '), 'Motivos externos: Corte de carretera')
  assert.equal(departureCancellationReason('minimum', 'No se alcanzo el minimo'), 'Falta de cupos: No se alcanzo el minimo')
})
