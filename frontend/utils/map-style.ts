type MapStyleLayer = {
  id: string
  type: string
  'source-layer'?: string
  paint?: Record<string, unknown>
}

export type MapStyle = {
  layers?: MapStyleLayer[]
  [key: string]: unknown
}

const palette = {
  background: '#080c0f',
  surface: '#111517',
  building: '#24292a',
  border: '#303432',
  roadMajor: '#f8f9f2',
  roadSecondary: '#b9baba',
  roadMinor: '#757b7a',
  roadLocal: '#4e5553',
  green: '#263d00',
  greenSoft: '#1b2b14',
  water: '#143642',
  waterLine: '#2f6677',
  sand: '#6c633c',
  accent: '#7d9505',
  primary: '#bafe06',
}

const setPaint = (layer: MapStyleLayer, property: string, value: unknown) => {
  layer.paint = { ...(layer.paint || {}), [property]: value }
}

export const buildAndariaMapStyle = (source: MapStyle): MapStyle => {
  const style = structuredClone(source)

  for (const layer of style.layers || []) {
    const sourceLayer = layer['source-layer'] || ''
    const id = layer.id.toLowerCase()

    if (layer.type === 'background') setPaint(layer, 'background-color', palette.background)

    if (layer.type === 'raster') {
      setPaint(layer, 'raster-brightness-max', 0.48)
      setPaint(layer, 'raster-saturation', -0.75)
      setPaint(layer, 'raster-contrast', 0.18)
    }

    if (sourceLayer === 'park') setPaint(layer, layer.type === 'line' ? 'line-color' : 'fill-color', palette.green)
    if (sourceLayer === 'landcover') {
      const color = id.includes('sand') ? palette.sand : id.includes('ice') ? palette.roadSecondary : id.includes('wetland') ? palette.waterLine : palette.greenSoft
      setPaint(layer, layer.type === 'line' ? 'line-color' : 'fill-color', color)
    }
    if (sourceLayer === 'landuse') {
      const color = id.includes('residential') ? palette.surface : id.includes('pitch') ? palette.green : palette.greenSoft
      setPaint(layer, layer.type === 'line' ? 'line-color' : 'fill-color', color)
    }
    if (sourceLayer === 'water') setPaint(layer, 'fill-color', palette.water)
    if (sourceLayer === 'waterway' && layer.type === 'line') setPaint(layer, 'line-color', palette.waterLine)
    if (sourceLayer === 'building') {
      if (layer.type === 'fill-extrusion') setPaint(layer, 'fill-extrusion-color', palette.building)
      else setPaint(layer, 'fill-color', palette.building)
    }
    if (sourceLayer === 'boundary') {
      setPaint(layer, 'line-color', palette.accent)
      setPaint(layer, 'line-opacity', id.includes('disputed') ? 0.55 : 0.72)
    }

    if (sourceLayer === 'transportation' && layer.type === 'line') {
      let color = palette.roadLocal
      if (id.includes('casing')) color = palette.background
      else if (id.includes('motorway') || id.includes('trunk') || id.includes('primary')) color = palette.roadMajor
      else if (id.includes('secondary') || id.includes('tertiary')) color = palette.roadSecondary
      else if (id.includes('minor') || id.includes('street') || id.includes('link')) color = palette.roadMinor
      else if (id.includes('path') || id.includes('pedestrian')) color = palette.accent
      setPaint(layer, 'line-color', color)
    }

    if (layer.type === 'symbol') {
      const isPlace = sourceLayer === 'place'
      const isRoad = sourceLayer === 'transportation_name'
      const isWater = sourceLayer === 'water_name' || sourceLayer === 'waterway'
      setPaint(layer, 'text-color', isPlace ? palette.roadMajor : isRoad ? palette.roadSecondary : isWater ? '#9bc9d5' : palette.roadSecondary)
      setPaint(layer, 'text-halo-color', palette.background)
      setPaint(layer, 'text-halo-width', isPlace ? 1.5 : 1)
      if (sourceLayer === 'poi') setPaint(layer, 'icon-opacity', 0.72)
    }
  }

  return style
}

export const andariaMapPalette = palette
