import assert from 'node:assert/strict'
import test from 'node:test'
import { navigationItems } from '../utils/navigation.ts'

test('guest navigation exposes the public home and both catalogs', () => {
  const paths = navigationItems('guest').map(item => item.to)
  assert.deepEqual(paths, ['/', '/attractions', '/packages'])
})

test('tourist navigation exposes packages and favorites without management links', () => {
  const paths = navigationItems('user', 'turista').map(item => item.to)
  assert.ok(paths.includes('/packages'))
  assert.ok(paths.includes('/app/favorites'))
  assert.ok(paths.includes('/app/purchases'))
  assert.ok(!paths.includes('/agency/packages'))
  assert.ok(!paths.includes('/managed-attractions'))
})

test('agency manager distinguishes public exploration from package management', () => {
  const items = navigationItems('user', 'encargado_agencia')
  assert.equal(items.find(item => item.to === '/packages')?.label, 'Explorar paquetes')
  assert.equal(items.find(item => item.to === '/agency/packages')?.label, 'Gestionar paquetes')
  assert.equal(items.find(item => item.to === '/agency/purchases')?.label, 'Revisar compras')
})
