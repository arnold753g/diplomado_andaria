# Registro de revisiones de Andaria

Antes de editar código o la monografía, abrir una revisión con `python docs/tools/registrar-revision.py begin --title "Descripción del cambio"`.
Si `docs/revisiones/activa.json` existe, continuar esa revisión sin sobrescribir su captura inicial.
Al finalizar las modificaciones y verificaciones, ejecutar `python docs/tools/registrar-revision.py close --result "Cambios, verificaciones y pendientes"`.
Usar el Python disponible en el entorno. El registro público está en `docs/revisiones/index.html`: nuevo a la izquierda en verde, anterior a la derecha en rojo.
No incluir credenciales, archivos de entorno reales ni datos personales en capturas o registros. No modificar fechas para simular un historial previo.
