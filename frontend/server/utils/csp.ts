export function nonceScripts(parts: string[], nonce: string): string[] {
  // Only apply to trusted renderer head/tail fragments, never the user-facing
  // SSR body. Text rendered from catalogue data must not gain script privileges.
  return parts.map(part => part.replace(/<script\b(?![^>]*\bnonce\s*=)/gi, `<script nonce="${nonce}"`))
}

export function contentSecurityPolicy(nonce: string, apiBase = ''): string {
  let apiOrigin = ''
  if (apiBase && !apiBase.startsWith('/')) {
    const url = new URL(apiBase)
    if (url.username || url.password || url.search || url.hash) throw new Error('The public API base must not contain credentials or query parameters')
    if (url.protocol !== 'https:' && !(url.protocol === 'http:' && ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname))) {
      throw new Error('The public API requires HTTPS')
    }
    apiOrigin = ` ${url.origin}`
  }
  return [
    "default-src 'self'", "base-uri 'self'", "object-src 'none'", "frame-ancestors 'none'",
    "form-action 'self' https://accounts.google.com",
    `script-src 'self' 'nonce-${nonce}'`,
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
    "font-src 'self' https://fonts.gstatic.com",
    "img-src 'self' data: blob: https://tile.openstreetmap.org https://tiles.openfreemap.org",
    `connect-src 'self' https://tiles.openfreemap.org${apiOrigin}`,
    "worker-src 'self' blob:", "upgrade-insecure-requests"
  ].join('; ')
}
