import { NextRequest, NextResponse } from 'next/server'
import { getApiBaseUrl, getApiHeaders } from '@/lib/api-config'

// GET /api/media - List media (proxies to backend)
export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const page = searchParams.get('page') ?? '1'
  const pageSize = searchParams.get('pageSize') ?? '20'
  const type = searchParams.get('type')

  const params = new URLSearchParams({ page, pageSize })
  if (type) params.set('type', type)

  try {
    const response = await fetch(`${getApiBaseUrl()}/media?${params}`, {
      headers: await getApiHeaders()
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('Proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to API' },
      { status: 502 }
    )
  }
}

// POST /api/media - Upload media (proxies to backend)
export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData()

    const response = await fetch(`${getApiBaseUrl()}/media/upload`, {
      method: 'POST',
      headers: await getApiHeaders(),
      body: formData
    })

    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('Proxy error:', error)
    return NextResponse.json(
      { error: 'Failed to connect to API' },
      { status: 502 }
    )
  }
}
