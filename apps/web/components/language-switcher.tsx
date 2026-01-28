'use client'

import { Languages } from 'lucide-react'
import { useTranslation } from '@/hooks/use-translations'
import { useLocale, type Locale } from '@/lib/i18n'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'

const languages: { code: Locale; label: string }[] = [
  { code: 'en', label: 'English' },
  { code: 'es', label: 'Español' }
]

export function LanguageSwitcher() {
  const { locale, setLocale } = useLocale()
  const { t } = useTranslation('language')

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant='outline' size='icon' title={t('switchLanguage')}>
          <Languages className='size-4' />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        {languages.map((lang) => (
          <DropdownMenuItem
            key={lang.code}
            onClick={() => setLocale(lang.code)}
            className={locale === lang.code ? 'bg-accent' : ''}
          >
            {lang.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
