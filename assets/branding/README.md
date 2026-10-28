# Iconos de Andaria

`andaria-icon-source.png` es la imagen proporcionada para el favicon web y el icono Android. Se conserva sin modificaciones; el generador elimina el margen transparente al exportar y mantiene transparencia, colores y proporciones, sin añadir fondos.

Desde la raíz, con `sharp` disponible en la resolución de módulos de Node:

```sh
node docs/tools/generar-iconos-andaria.cjs
```

Si se usa el runtime de Codex, establecer `NODE_PATH` al directorio de paquetes indicado por `load_workspace_dependencies` antes de ejecutar el comando. No se incorpora esa ruta local al proyecto ni se necesita ejecutar el generador durante el despliegue: los recursos exportados se incluyen en el repositorio.

Salidas: favicon SVG e ICO, iconos táctiles y Flutter web, iconos Android por densidad y primer plano adaptativo de 108dp. El símbolo de Android ocupa 46dp para quedar dentro del círculo seguro de 66dp y conservar los cuatro cuadrantes con máscaras redondas.

El símbolo transparente también se exporta a `mobile/assets/branding/andaria-symbol.png` para el encabezado y la carga animada de Flutter; la animación utiliza el mismo archivo, sin recrear ni recolorear la marca.
