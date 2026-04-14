/** Decode JWT payload (no signature verification; display-only). */
export function parseJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split('.')
    if (parts.length < 2 || !parts[1]) {
      return null
    }
    const b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const pad = b64.length % 4 === 0 ? '' : '='.repeat(4 - (b64.length % 4))
    const json = atob(b64 + pad)
    return JSON.parse(json) as Record<string, unknown>
  } catch {
    return null
  }
}

export function jwtSubject(token: string | null): string | null {
  if (!token) {
    return null
  }
  const p = parseJwtPayload(token)
  const sub = p?.sub
  return typeof sub === 'string' ? sub : null
}
