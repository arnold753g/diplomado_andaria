# Registro de cambios

Abrir `index.html` en el navegador. Cada revisión reúne los cambios de una tarea, identifica su fecha y permite buscar archivos y comparar fragmentos. La izquierda muestra lo modificado en verde y la derecha lo anterior en rojo; los números remiten a las líneas de cada versión.

Antes de cambiar fuentes, ejecutar desde la raíz del proyecto:

```powershell
python docs/tools/registrar-revision.py begin --title "Descripción de la revisión"
```

Después de editar y verificar:

```powershell
python docs/tools/registrar-revision.py close --result "Cambios, pruebas y pendientes"
```

Para actualizar la vista de una revisión abierta: `python docs/tools/registrar-revision.py render`.
Las capturas `antes.json` y `despues.json` permiten reconstruir la comparación. Mantenerlas con Git junto con las fuentes: este archivo registra revisiones; no reemplaza los commits ni detecta automáticamente modificaciones hechas fuera de este procedimiento.

Se comparan fuentes de la monografía, backend, proxy y documentación pública. Las nuevas revisiones también capturan Vue/TypeScript/CSS, Dart, configuración pública Android/iOS y migraciones SQL. La primera revisión conserva la cobertura con la que se abrió: no presenta archivos intactos como incorporaciones. Se excluyen evidencias de sesiones, archivos de entorno, archivos generados y credenciales. Los PDF e imágenes se verifican mediante sus informes y manifiestos, sin duplicarlos en el registro. Las correcciones previas al inicio de este registro permanecen en sus informes históricos.
