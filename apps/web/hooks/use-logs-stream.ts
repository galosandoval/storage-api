'use client'

import { useCallback, useEffect, useRef, useState, useEffectEvent } from 'react'

export interface LogEntry {
  id: string
  message: string
  timestamp: Date
  priority: number
  pid: string
}

interface RawLogMessage {
  MESSAGE?: string
  __REALTIME_TIMESTAMP?: string
  PRIORITY?: string
  _PID?: string
}

interface UseLogsStreamOptions {
  lineCount?: number
  since?: string
}

interface UseLogsStreamReturn {
  logs: LogEntry[]
  isConnected: boolean
  error: string | null
  clearLogs: () => void
}

/**
 * Get WebSocket URL for logs stream.
 * Uses NEXT_PUBLIC_PI_WS_URL or constructs from NEXT_PUBLIC_PI_HOST + NEXT_PUBLIC_PI_PORT
 */
function getLogsWsUrl(params: URLSearchParams): string | null {
  const host = process.env.NEXT_PUBLIC_PI_HOST
  if (!host) {
    return null
  }

  const port = process.env.NEXT_PUBLIC_PI_PORT || '8080'
  return `ws://${host}:${port}/logs/stream?${params}`
}

export function useLogsStream(
  options: UseLogsStreamOptions = {}
): UseLogsStreamReturn {
  const { lineCount = 50, since } = options
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const setErrorEvent = useEffectEvent((error: string) => {
    setError(error)
  })
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const idCounterRef = useRef(0)

  const clearLogs = useCallback(() => {
    setLogs([])
  }, [])

  useEffect(() => {
    const params = new URLSearchParams({ n: String(lineCount) })
    if (since) params.set('since', since)

    const wsUrl = getLogsWsUrl(params)
    if (!wsUrl) {
      console.error(
        'WebSocket URL not configured. Set NEXT_PUBLIC_PI_HOST in .env and restart the dev server.',
        { NEXT_PUBLIC_PI_HOST: process.env.NEXT_PUBLIC_PI_HOST }
      )
      setErrorEvent('Set NEXT_PUBLIC_PI_HOST in .env and restart dev server')
      return
    }

    function connect(url: string) {
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        setIsConnected(true)
        setError(null)
      }

      ws.onmessage = (event) => {
        try {
          const raw: RawLogMessage = JSON.parse(event.data)
          const entry: LogEntry = {
            id: `log-${idCounterRef.current++}`,
            message: raw.MESSAGE ?? '',
            timestamp: new Date(Number(raw.__REALTIME_TIMESTAMP ?? '0') / 1000),
            priority: Number(raw.PRIORITY ?? '6'),
            pid: raw._PID ?? ''
          }
          setLogs((prev) => [...prev, entry])
        } catch {
          // Skip malformed messages
        }
      }

      ws.onerror = () => {
        setError('WebSocket connection error')
      }

      ws.onclose = () => {
        setIsConnected(false)
        // Attempt reconnect after 3 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          connect(url)
        }, 3000)
      }
    }

    connect(wsUrl)

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [lineCount, since])

  return { logs, isConnected, error, clearLogs }
}
