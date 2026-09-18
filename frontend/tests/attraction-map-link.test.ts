import assert from 'node:assert/strict'
import test from 'node:test'
import { googleMapsDirectionsURL } from '../utils/attractions.ts'

test('creates a mobile route with the exact coordinates as destination', () => {
  assert.equal(
    googleMapsDirectionsURL(-21.535486, -64.729557),
    'https://www.google.com/maps/dir/?api=1&destination=-21.535486%2C-64.729557&travelmode=driving'
  )
})
