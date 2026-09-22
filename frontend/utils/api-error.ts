import type { ApiErrorData } from '~/types/api'

export const apiError = (error: unknown, fallback = 'No se pudo completar la operación'): ApiErrorData => {
  const value = error as { data?: { error?: ApiErrorData } }
  const result = value?.data?.error
  if (!result) return { code: 'REQUEST_FAILED', message: fallback }
  const messages: Record<string, string> = {
    INVALID_CREDENTIALS: 'El correo o la contraseña no son correctos.',
    ACCOUNT_LOCKED: 'Demasiados intentos fallidos. Espera unos minutos antes de volver a intentar.',
    ACCOUNT_INACTIVE: 'Tu cuenta está desactivada. Contacta al administrador.',
    EMAIL_EXISTS: 'Ya existe una cuenta con ese correo.',
    PASSWORD_POLICY: 'Usa una contraseña de entre 12 y 72 bytes, sin espacios al inicio o al final.',
    PASSWORD_REUSED: 'La nueva contraseña debe ser diferente de la actual.',
    VALIDATION_ERROR: 'Revisa los datos ingresados y sus límites de longitud.',
    REGISTRATION_DISABLED: 'El registro de cuentas nuevas está deshabilitado.',
    FORBIDDEN: 'No tienes permiso para realizar esta acción.',
    SELF_ADMIN_CHANGE: 'Modifica tus datos desde Mi perfil; no puedes cambiar tu propio rol o estado.',
    INTERNAL_ERROR: 'No se pudo completar la operación. Inténtalo nuevamente.',
    INVALID_JSON: 'Los datos enviados no son válidos.',
    CSRF_INVALID: 'Tu sesión necesita actualizarse. Recarga la página y vuelve a intentar.',
    DEPARTURE_NOT_BOOKABLE: 'La salida ya no está disponible para compra.',
    CAPACITY_UNAVAILABLE: 'Ya no hay cupos suficientes para todo el grupo.',
    PAYMENT_METHOD_UNAVAILABLE: 'La agencia ya no admite ese medio de pago.',
    PURCHASE_CONFLICT: 'La compra cambió. Recarga la página antes de continuar.',
    PURCHASE_STATE_CONFLICT: 'La compra ya fue revisada o cambió de estado.',
    PURCHASE_PROOF_INVALID: 'Sube un comprobante PNG o JPG de hasta 5 MB.'
  }
  return { ...result, message: messages[result.code] || result.message }
}
