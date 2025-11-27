import { useState, useEffect } from 'react'
import { Loader2, Monitor, AlertCircle, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

type InstanceStatus = 'pending' | 'running' | 'terminating'

interface User {
  id: string
  type: 'openid_connect' | 'anonymous'
}

interface AuthStatus {
  user: User
  instance?: {
    status: InstanceStatus
    pod_ip?: string
  }
}

function App() {
  const [state, setState] = useState<'checking' | 'idle' | 'pending' | 'error'>('checking')
  const [errorMsg, setErrorMsg] = useState<string>('')
  const [status, setStatus] = useState<string>('')

  const checkAuth = async () => {
    try {
      const res = await fetch('/_hakoniwa/api/auth/me')
      if (res.status === 401) {
        setState('idle')
        return
      }
      if (!res.ok) {
        throw new Error('Failed to check auth')
      }
      const data: AuthStatus = await res.json()
      handleAuthData(data)
    } catch (err) {
      console.error(err)
      setState('error')
      setErrorMsg('Failed to connect to server')
    }
  }

  const handleAuthData = (data: AuthStatus) => {
    const instanceStatus = data.instance?.status
    setStatus(instanceStatus || 'none')

    if (instanceStatus === 'running') {
      window.location.reload()
      return
    }

    if (instanceStatus === 'pending' || instanceStatus === 'terminating') {
      setState('pending')
    } else {
      setState('idle')
    }
  }

  const loginAnonymous = async () => {
    setState('checking')
    try {
      const res = await fetch('/_hakoniwa/api/auth/anonymous', { method: 'POST' })
      if (!res.ok) throw new Error('Login failed')
      const data: AuthStatus = await res.json()
      handleAuthData(data)
    } catch (err) {
      console.error(err)
      setState('error')
      setErrorMsg('Login failed')
    }
  }

  useEffect(() => {
    checkAuth()
  }, [])

  useEffect(() => {
    let interval: number
    if (state === 'pending') {
      interval = setInterval(checkAuth, 3000)
    }
    return () => clearInterval(interval)
  }, [state])

  if (state === 'checking') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="h-10 w-10 animate-spin text-primary" />
          <p className="text-muted-foreground">Connecting to Hakoniwa...</p>
        </div>
      </div>
    )
  }

  if (state === 'pending') {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle className="text-2xl">Preparing Desktop</CardTitle>
            <CardDescription>We are spinning up your personal environment.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-6 py-6">
            <Loader2 className="h-16 w-16 animate-spin text-primary" />
            <div className="text-sm text-muted-foreground">
              Status: <span className="font-mono font-medium text-foreground">{status}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (state === 'error') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md border-destructive">
          <CardHeader>
            <div className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-6 w-6" />
              <CardTitle>Error Occurred</CardTitle>
            </div>
            <CardDescription>Something went wrong while connecting.</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm">{errorMsg}</p>
          </CardContent>
          <CardFooter>
            <Button variant="destructive" className="w-full" onClick={() => window.location.reload()}>
              Retry
            </Button>
          </CardFooter>
        </Card>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-500 to-purple-600 p-4">
      <Card className="w-full max-w-lg shadow-2xl">
        <CardHeader className="text-center space-y-2">
          <div className="flex justify-center mb-2">
            <div className="p-3 bg-primary/10 rounded-full">
              <Monitor className="h-10 w-10 text-primary" />
            </div>
          </div>
          <CardTitle className="text-3xl font-bold">Hakoniwa</CardTitle>
          <CardDescription className="text-lg">
            On-Demand Cloud Desktop Environment
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          <Button variant="outline" className="w-full h-12 text-lg" disabled>
            Login with OIDC (Coming Soon)
          </Button>
          
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background px-2 text-muted-foreground">Or continue with</span>
            </div>
          </div>

          <Button 
            className="w-full h-12 text-lg group" 
            onClick={loginAnonymous}
          >
            Try Anonymously
            <ArrowRight className="ml-2 h-4 w-4 transition-transform group-hover:translate-x-1" />
          </Button>
        </CardContent>
        <CardFooter className="justify-center">
          <p className="text-xs text-muted-foreground text-center">
            By continuing, you agree to our Terms of Service and Privacy Policy.
          </p>
        </CardFooter>
      </Card>
    </div>
  )
}

export default App
