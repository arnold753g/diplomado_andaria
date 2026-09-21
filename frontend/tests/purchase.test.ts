import assert from 'node:assert/strict'
import test from 'node:test'
import { paymentMethodLabel, purchaseSeverity, purchaseStatus, refundMethodLabel } from '../utils/purchases.ts'

test('purchase states distinguish correction from refund', () => {
  assert.equal(purchaseStatus('payment_review'), 'Pago en revisión')
  assert.equal(purchaseStatus('correction_requested'), 'Corrección solicitada')
  assert.equal(purchaseStatus('refund_pending'), 'Reembolso pendiente')
  assert.equal(purchaseSeverity('confirmed'), 'success')
  assert.equal(purchaseSeverity('refund_pending'), 'warn')
})

test('payment methods use tourist-facing labels', () => {
  assert.equal(paymentMethodLabel('qr'), 'Pago mediante QR')
  assert.equal(paymentMethodLabel('transfer'), 'Transferencia bancaria')
})

test('refund destinations distinguish QR, bank and missing data', () => {
  assert.equal(refundMethodLabel('qr'), 'QR para devolución')
  assert.equal(refundMethodLabel('bank_transfer'), 'Cuenta bancaria')
  assert.equal(refundMethodLabel(), 'Pendiente de datos del turista')
})
