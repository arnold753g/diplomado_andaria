import type { PaymentMethod, PurchaseStatus } from '../types/purchase.ts'

export const purchaseStatus = (status: PurchaseStatus) => ({
  payment_review: 'Pago en revisión',
  correction_requested: 'Corrección solicitada',
  confirmed: 'Compra confirmada',
  payment_rejected: 'Pago rechazado',
  cancelled: 'Cancelada',
  refund_pending: 'Reembolso pendiente',
  refunded: 'Reembolsada'
}[status])

export const purchaseSeverity = (status: PurchaseStatus) => ({
  payment_review: 'warn', correction_requested: 'info', confirmed: 'success', payment_rejected: 'danger', cancelled: 'secondary', refund_pending: 'warn', refunded: 'info'
}[status] as 'warn' | 'success' | 'danger' | 'secondary' | 'info')

export const paymentMethodLabel = (method: PaymentMethod) => method === 'qr' ? 'Pago mediante QR' : 'Transferencia bancaria'

export const refundMethodLabel = (method?: 'qr' | 'bank_transfer') => method === 'qr' ? 'QR para devolución' : method === 'bank_transfer' ? 'Cuenta bancaria' : 'Pendiente de datos del turista'
