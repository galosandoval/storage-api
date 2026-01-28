import { parseAsStringLiteral } from 'nuqs'

export const typeFilterParser = parseAsStringLiteral([
  'all',
  'photo',
  'video'
] as const).withDefault('all')

export const visibilityFilterParser = parseAsStringLiteral([
  'all',
  'mine',
  'public'
] as const).withDefault('all')
