'use client'

import { useTranslations as useNextIntlTranslations } from 'next-intl'
import type messages from '@/messages/en.json'

// Infer the message structure from the JSON
type Messages = typeof messages

// Get top-level namespace keys (e.g., "common", "header", "upload")
type Namespace = keyof Messages

// Recursively build dot-notation paths for nested objects
type NestedPaths<T, Prefix extends string = ''> = T extends string
  ? Prefix extends ''
    ? never
    : Prefix
  : {
      [K in keyof T & string]: NestedPaths<
        T[K],
        Prefix extends '' ? K : `${Prefix}.${K}`
      >
    }[keyof T & string]

// All possible translation keys as dot-notation strings
type TranslationKey = NestedPaths<Messages>

// Get keys within a specific namespace
type NamespacedKey<N extends Namespace> = keyof Messages[N] & string

// Translation function type - returns the translated string
type TranslateFunction = (key: TranslationKey) => string

// Namespaced translation function type
type NamespacedTranslateFunction<N extends Namespace> = (
  key: NamespacedKey<N>
) => string

/**
 * Type-safe translation hook with full autocomplete support.
 *
 * Usage without namespace (full dot-notation keys):
 * ```
 * const { t } = useTranslation()
 * t('common.signOut') // autocomplete works!
 * ```
 *
 * Usage with namespace (shorter keys):
 * ```
 * const { t } = useTranslation('common')
 * t('signOut') // autocomplete works!
 * ```
 */
export function useTranslation(): { t: TranslateFunction }
export function useTranslation<N extends Namespace>(
  namespace: N
): { t: NamespacedTranslateFunction<N> }
export function useTranslation<N extends Namespace>(namespace?: N) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const t = useNextIntlTranslations(namespace) as any

  return {
    t: (key: string) => t(key) as string
  }
}

// Re-export types for external use
export type { TranslationKey, Namespace, Messages }
