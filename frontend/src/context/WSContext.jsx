import { createContext, useContext, useEffect, useRef, useState } from 'react'
import { useAuth } from './AuthContext'

const WSContext = createContext(null)

export function WSProvider({ children }) {
  const { token } = useAuth()
  const [metrics, setMetrics] = useState(null)
  const [connected, setConnected] = useState(false)
  const ws = useRef(null)
  const reconnectTimeout = useRef(null)

  const connect = () => {
    console.log('connect called, token:', token)
    if (!token) return

    ws.current = new WebSocket(`ws://localhost:8080/api/v1/ws?token=${token}`)

    ws.current.onopen = () => {
      setConnected(true)
      console.log('WS connected')
    }

    ws.current.onmessage = (e) => {
      const data = JSON.parse(e.data)
      setMetrics(data)
    }

    ws.current.onclose = () => {
      setConnected(false)
      console.log('WS disconnected, retrying in 3s...')
      reconnectTimeout.current = setTimeout(connect, 3000)
    }

    ws.current.onerror = (err) => {
      console.error('WS error:', err)
      ws.current.close()
    }
  }

  useEffect(() => {
    connect()
    return () => {
      clearTimeout(reconnectTimeout.current)
      ws.current?.close()
    }
  }, [token])

  return (
    <WSContext.Provider value={{ metrics, connected }}>
      {children}
    </WSContext.Provider>
  )
}

export function useWS() {
  return useContext(WSContext)
}