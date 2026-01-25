// Server-side only configuration for the backend API
// These are not exposed to the browser

/**
 * Get the base URL for the backend API.
 *
 * Configure via environment variables:
 * - PI_API_URL: Full URL (e.g., "http://192.168.1.100:8080")
 * - Or use PI_HOST + PI_PORT separately
 *
 * Defaults to http://localhost:8080 if not configured.
 */
export function getApiBaseUrl(): string {
  // Option 1: Full URL provided
  if (process.env.PI_API_URL) {
    return process.env.PI_API_URL.replace(/\/$/, '') // Remove trailing slash
  }
  console.log('process.env', process.env)
  // Option 2: Host + Port (for backward compatibility with NEXT_PUBLIC_PI_HOST)
  const host =
    process.env.PI_HOST || process.env.NEXT_PUBLIC_PI_HOST || 'localhost'
  const port = process.env.PI_PORT || '8080'

  return `http://${host}:${port}`
}

/**
 * Get the household ID for API requests.
 * This is scaffolding for future auth integration.
 */
export function getHouseholdId(): string {
  return process.env.HOUSEHOLD_ID || process.env.NEXT_PUBLIC_HOUSEHOLD_ID || ''
}

/**
 * Get common headers for backend API requests.
 */
export function getApiHeaders(): HeadersInit {
  return {
    'X-Household-ID': getHouseholdId()
  }
}
