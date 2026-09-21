# Auditoría funcional del módulo 6: cancelaciones y reembolsos

Fecha de revisión: 23 de septiembre de 2026.

## Resultado

El módulo quedó aprobado en los 22 casos ejecutados. La evaluación del cupo mínimo y la decisión comercial de cancelar son operaciones separadas: el proceso automático deja la salida en `minimum_review` y únicamente el encargado de la agencia puede iniciar después los reembolsos.

La auditoría cubrió:

- cancelación automática del turista dentro de la política;
- rechazo fuera del plazo sin modificar cupos;
- liberación transaccional de capacidad retenida o confirmada;
- vencimiento exacto de 72 horas;
- destino por QR y por cuenta bancaria;
- validación de datos incompletos y comprobante obligatorio;
- registro de referencia, fecha y responsable de la devolución;
- acceso aislado a QR y comprobantes;
- permisos de visitante, turista, administrador y ambos tipos de encargado;
- decisión de cupo mínimo limitada a la agencia asignada;
- control de versiones ante acciones repetidas;
- restricciones persistentes en PostgreSQL;
- bandejas separadas para turista y agencia;
- pruebas Go, análisis estático, pruebas frontend, tipos y compilación.

## Límites conocidos

La transferencia real continúa fuera del sistema y no existe conciliación bancaria automática. Tampoco se envían todavía recordatorios por correo por vencimientos de reembolso. La interfaz autenticada se validó mediante API, pruebas de lógica, tipos y compilación; la revisión visual de este ciclo se realizó en el catálogo público.

No se realizaron pruebas de carga ni inyección de fallos. Estas acciones están deshabilitadas en la configuración de Medusa para el entorno de desarrollo.

La matriz reproducible y el resultado detallado están en [la matriz](../.medusa-auditor/results/refunds-matrix.json) y [el reporte de Medusa](../.medusa-auditor/results/refunds-report.md).
