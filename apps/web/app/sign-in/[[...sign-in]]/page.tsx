import { SignIn, SignOutButton } from '@clerk/nextjs'
import { auth } from '@clerk/nextjs/server'
import { AlertCircle } from 'lucide-react'

interface SignInPageProps {
  searchParams: Promise<{ error?: string }>
}

export default async function SignInPage({ searchParams }: SignInPageProps) {
  const { error } = await searchParams
  const { userId } = await auth()

  const isUnauthorized = error === 'unauthorized'

  // If user is signed in but unauthorized, show sign out option
  if (isUnauthorized && userId) {
    return (
      <div className='min-h-svh flex items-center justify-center bg-background'>
        <div className='flex flex-col items-center gap-6 max-w-md text-center px-4'>
          <div className='flex items-center gap-2 text-destructive'>
            <AlertCircle className='size-6' />
            <h1 className='text-xl font-semibold'>Access Denied</h1>
          </div>
          <p className='text-muted-foreground'>
            You are not authorized to access this application. Only household
            members can sign in.
          </p>
          <SignOutButton>
            <button className='px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors'>
              Sign out and try a different account
            </button>
          </SignOutButton>
        </div>
      </div>
    )
  }

  return (
    <div className='min-h-svh flex items-center justify-center bg-background'>
      <div className='flex flex-col items-center gap-6'>
        {isUnauthorized && (
          <div className='flex items-center gap-2 text-destructive bg-destructive/10 px-4 py-3 rounded-lg'>
            <AlertCircle className='size-5' />
            <p className='text-sm'>
              You are not authorized to access this application.
            </p>
          </div>
        )}
        <SignIn />
      </div>
    </div>
  )
}
