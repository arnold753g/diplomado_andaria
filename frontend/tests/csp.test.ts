import assert from 'node:assert/strict'
import test from 'node:test'
import { contentSecurityPolicy, nonceScripts } from '../server/utils/csp.ts'

test('CSP authorizes the renderer nonce without allowing arbitrary inline scripts', () => {
  const policy = contentSecurityPolicy('test-nonce')
  const scripts = policy.split('; ').find(part => part.startsWith('script-src'))!
  assert.equal(scripts, "script-src 'self' 'nonce-test-nonce'")
  assert.ok(policy.includes("object-src 'none'"))
  assert.ok(policy.includes("frame-ancestors 'none'"))
})

test('trusted head and tail scripts receive nonces without rewriting script content', () => {
  const parts = ['<script type="module">console.log("ok")</script>', '<link rel="stylesheet">', '<script src="/app.js"></script>']
  assert.deepEqual(nonceScripts(parts, 'abc'), [
    '<script nonce="abc" type="module">console.log("ok")</script>', '<link rel="stylesheet">', '<script nonce="abc" src="/app.js"></script>'
  ])
  assert.equal(parts[0], '<script type="module">console.log("ok")</script>')
})

test('CSP permits a configured HTTPS API without allowing arbitrary HTTP hosts', () => {
  assert.ok(contentSecurityPolicy('abc', 'https://api.example.test/api/v1').includes('https://api.example.test'))
  assert.throws(() => contentSecurityPolicy('abc', 'http://api.example.test/api/v1'))
  assert.ok(contentSecurityPolicy('abc', '/api/v1').includes("connect-src 'self'"))
})
