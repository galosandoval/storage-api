import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <div className='min-h-svh flex items-center justify-center bg-background'>
      <SignIn />
    </div>
  )
}
