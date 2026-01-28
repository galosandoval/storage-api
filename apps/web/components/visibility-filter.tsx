'use client'

import { Globe, User } from 'lucide-react'
import { useTranslation } from '@/hooks/use-translations'
import type { VisibilityFilter } from '@/lib/types/media'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

interface VisibilityFilterProps {
  value: VisibilityFilter
  onChange: (value: VisibilityFilter) => void
}

export function VisibilityFilter({ value, onChange }: VisibilityFilterProps) {
  const { t } = useTranslation('filters')

  return (
    <ToggleGroup
      type='single'
      value={value}
      onValueChange={(val) => {
        if (val) {
          onChange(val as VisibilityFilter)
        }
      }}
      className='justify-start'
    >
      <ToggleGroupItem value='all' aria-label={t('showAll')}>
        <Globe className='size-4 mr-2' />
        {t('all')}
      </ToggleGroupItem>
      <ToggleGroupItem value='mine' aria-label={t('showMine')}>
        <User className='size-4 mr-2' />
        {t('mine')}
      </ToggleGroupItem>
    </ToggleGroup>
  )
}
