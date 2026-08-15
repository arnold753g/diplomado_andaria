import assert from 'node:assert/strict'
import test from 'node:test'
import { validPassword } from '../utils/password.ts'
test('password policy measures UTF-8 bytes and rejects boundary whitespace', () => {
  assert.equal(validPassword('a'.repeat(12)), true)
  assert.equal(validPassword('a'.repeat(11)), false)
  assert.equal(validPassword('á'.repeat(36)), true)
  assert.equal(validPassword('á'.repeat(37)), false)
  assert.equal(validPassword(' a'.repeat(12)), false)
})
