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

/** `uid` claim from Kong Manager access tokens (local or SSO-linked user). */
export function jwtUserId(token: string | null): number | null {
  if (!token) {
    return null
  }
  const p = parseJwtPayload(token)
  const uid = p?.uid
  if (typeof uid === 'number' && Number.isFinite(uid)) {
    return uid
  }
  if (typeof uid === 'string' && /^\d+$/.test(uid)) {
    return parseInt(uid, 10)
  }
  return null
}

/** Session expiry from JWT `exp` (seconds since epoch). */
export function jwtExpiresAt(token: string | null): Date | null {
  if (!token) {
    return null
  }
  const p = parseJwtPayload(token)
  const exp = p?.exp
  if (typeof exp === 'number' && Number.isFinite(exp)) {
    return new Date(exp * 1000)
  }
  return null
}
