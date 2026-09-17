import test from 'node:test'
import assert from 'node:assert/strict'
import { buildAndariaMapStyle } from '../utils/map-style.ts'

test('Andaria map style recolors functional layers without mutating the base style', () => {
  const base = {
    version: 8,
    layers: [
      { id: 'background', type: 'background', paint: { 'background-color': '#ffffff' } },
      { id: 'water', type: 'fill', 'source-layer': 'water', paint: { 'fill-color': '#0000ff' } },
      { id: 'road_trunk_primary', type: 'line', 'source-layer': 'transportation', paint: { 'line-color': '#ff0000' } },
      { id: 'label_city', type: 'symbol', 'source-layer': 'place', paint: { 'text-color': '#000000' } },
      { id: 'waterway_line_label', type: 'symbol', 'source-layer': 'waterway', paint: { 'text-color': '#0000ff' } },
    ],
  }

  const styled = buildAndariaMapStyle(base)
  assert.notEqual(styled, base)
  assert.equal(base.layers[0]?.paint?.['background-color'], '#ffffff')
  assert.equal(styled.layers?.[0]?.paint?.['background-color'], '#080c0f')
  assert.equal(styled.layers?.[1]?.paint?.['fill-color'], '#143642')
  assert.equal(styled.layers?.[2]?.paint?.['line-color'], '#f8f9f2')
  assert.equal(styled.layers?.[3]?.paint?.['text-halo-color'], '#080c0f')
  assert.equal(styled.layers?.[4]?.paint?.['line-color'], undefined)
  assert.equal(styled.layers?.[4]?.paint?.['text-color'], '#9bc9d5')
})
