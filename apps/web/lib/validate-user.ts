import { redirect } from 'next/navigation'
import { auth } from '@clerk/nextjs/server'
import { getApiBaseUrl, getApiHeaders } from './api-config'

export interface ValidatedUser {
  id: string
  householdId: string
  email: string
  firstName?: string
  lastName?: string
  imageUrl?: string
  role: string
}

interface MeResponse {
  user?: ValidatedUser
  error?: string
}

/**
 * Validates that the current Clerk user is authorized for this household.
 * If not authorized, redirects to sign-in with error.
 * The sign-in page will handle signing them out.
 *
 * Call this at the top of protected server components.
 */
export async function validateUser(): Promise<ValidatedUser> {
  const { userId } = await auth()

  if (!userId) {
    redirect('/sign-in')
  }

  try {
    const response = await fetch(`${getApiBaseUrl()}/me`, {
      headers: await getApiHeaders(),
      cache: 'no-store'
    })

    if (!response.ok) {
      if (response.status === 401 || response.status === 403) {
        // User not in household - redirect to sign-in with error
        redirect('/sign-in?error=unauthorized')
      }
      throw new Error(`API error: ${response.status}`)
    }

    const data: MeResponse = await response.json()

    if (!data.user) {
      redirect('/sign-in?error=unauthorized')
    }

    return data.user
  } catch (error) {
    // If it's a redirect, let it propagate
    if (
      error &&
      typeof error === 'object' &&
      'digest' in error &&
      typeof (error as { digest: unknown }).digest === 'string' &&
      (error as { digest: string }).digest.startsWith('NEXT_REDIRECT')
    ) {
      throw error
    }

    console.error('User validation failed:', error)
    // On API errors, redirect to sign-in for safety
    redirect('/sign-in?error=unauthorized')
  }
}
