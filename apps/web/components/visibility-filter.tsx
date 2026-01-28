'use client'

import { Globe, User } from 'lucide-react'
import type { VisibilityFilter } from '@/lib/types/media'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

interface VisibilityFilterProps {
  value: VisibilityFilter
  onChange: (value: VisibilityFilter) => void
}

export function VisibilityFilter({ value, onChange }: VisibilityFilterProps) {
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
      <ToggleGroupItem value='all' aria-label='Show all'>
        <Globe className='size-4 mr-2' />
        All
      </ToggleGroupItem>
      <ToggleGroupItem value='mine' aria-label='Show my uploads'>
        <User className='size-4 mr-2' />
        Mine
      </ToggleGroupItem>
    </ToggleGroup>
  )
}
