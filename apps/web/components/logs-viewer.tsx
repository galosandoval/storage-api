'use client'

import { useEffect, useRef, useState } from 'react'
import { Circle, Trash2 } from 'lucide-react'
import { useLogsStream, type LogEntry } from '@/hooks/use-logs-stream'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'

function formatTimestamp(date: Date): string {
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

function LogLine({ entry }: { entry: LogEntry }) {
  return (
    <div className='flex gap-3 px-4 py-0.5 hover:bg-white/5 transition-colors'>
      <span className='text-emerald-400/70 shrink-0 tabular-nums'>
        {formatTimestamp(entry.timestamp)}
      </span>
      <span className='text-muted-foreground shrink-0 tabular-nums w-16'>
        [{entry.pid}]
      </span>
      <span className='text-foreground break-all'>{entry.message}</span>
    </div>
  )
}

export function LogsViewer() {
  const { logs, isConnected, error, clearLogs } = useLogsStream({
    lineCount: 100
  })
  const scrollRef = useRef<HTMLDivElement>(null)
  const [autoScroll, setAutoScroll] = useState(true)

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      const viewport = scrollRef.current.querySelector(
        '[data-radix-scroll-area-viewport]'
      )
      if (viewport) {
        viewport.scrollTop = viewport.scrollHeight
      }
    }
  }, [logs, autoScroll])

  // Detect manual scroll to disable auto-scroll
  const handleScroll = (event: React.UIEvent<HTMLDivElement>) => {
    const target = event.currentTarget.querySelector(
      '[data-radix-scroll-area-viewport]'
    )
    if (!target) return

    const { scrollTop, scrollHeight, clientHeight } = target
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50
    setAutoScroll(isAtBottom)
  }

  return (
    <Card className='w-full max-w-5xl bg-card border-border shadow-2xl'>
      <CardHeader className='flex flex-row items-center justify-between border-b border-border py-3'>
        <div className='flex items-center gap-3'>
          <CardTitle className='text-foreground text-base font-medium tracking-tight'>
            System Logs
          </CardTitle>
          <div className='flex items-center gap-1.5'>
            <Circle
              className={cn(
                'size-2 fill-current',
                isConnected ? 'text-emerald-500' : 'text-red-500'
              )}
            />
            <span className='text-xs text-muted-foreground'>
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
        </div>

        <div className='flex items-center gap-2'>
          {!autoScroll && (
            <Button
              variant='ghost'
              size='sm'
              onClick={() => setAutoScroll(true)}
              className='text-xs text-muted-foreground hover:text-foreground hover:bg-accent'
            >
              Resume auto-scroll
            </Button>
          )}
          <Button
            variant='ghost'
            size='sm'
            onClick={clearLogs}
            className='text-muted-foreground hover:text-foreground hover:bg-accent'
          >
            <Trash2 className='size-4' />
          </Button>
        </div>
      </CardHeader>

      <CardContent className='p-0'>
        {error ? (
          <div className='p-4 text-red-400 text-sm font-mono'>{error}</div>
        ) : (
          <div ref={scrollRef} onScroll={handleScroll}>
            <ScrollArea className='h-[600px]'>
              <div className='font-mono text-sm py-2'>
                {logs.length === 0 ? (
                  <div className='px-4 py-8 text-muted-foreground text-center'>
                    Waiting for logs...
                  </div>
                ) : (
                  logs.map((entry) => <LogLine key={entry.id} entry={entry} />)
                )}
              </div>
            </ScrollArea>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
