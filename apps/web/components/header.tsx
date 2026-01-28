'use client'

import { SignOutButton, UserButton } from '@clerk/nextjs'
import { LogOut } from 'lucide-react'
import { useTranslation } from '@/hooks/use-translations'
import { Button } from '@/components/ui/button'
import { LanguageSwitcher } from '@/components/language-switcher'

export function Header() {
  const { t } = useTranslation()

  return (
    <header className='flex items-center justify-between p-4 sm:px-8 border-b'>
      <h1 className='text-xl font-semibold'>{t('header.title')}</h1>
      <div className='flex items-center gap-3'>
        <LanguageSwitcher />
        <SignOutButton>
          <Button variant='outline' size='sm'>
            <LogOut className='size-4 mr-2' />
            {t('common.signOut')}
          </Button>
        </SignOutButton>
        <UserButton />
      </div>
    </header>
  )
}
