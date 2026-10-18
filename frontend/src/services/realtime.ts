import { ref } from 'vue'
import { refreshSession } from '../api'
export const connected = ref(false)
type Event = { type: string; request_id?: string; data?: unknown }
const listeners = new Set<(event: Event) => void>()
let socket: WebSocket | null = null,
  active = false,
  timer: ReturnType<typeof setTimeout> | undefined,
  attempt = 0
export function subscribe(listener: (event: Event) => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}
export function startRealtime() {
  active = true
  connect()
}
function connect() {
  if (!active || socket) return
  const ws = new WebSocket(
    `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/api/v1/ws`,
  )
  socket = ws
  ws.onopen = () => {
    connected.value = true
    attempt = 0
  }
  ws.onmessage = (message) => {
    let event: Event
    try {
      event = JSON.parse(message.data)
    } catch {
      return
    }
    for (const listener of listeners) listener(event)
  }
  ws.onclose = () => {
    if (socket !== ws) return
    socket = null
    connected.value = false
    if (active)
      timer = setTimeout(
        async () => {
          try {
            await refreshSession()
            if (active) connect()
          } catch {
            window.dispatchEvent(new Event('squadflow:session-expired'))
          }
        },
        Math.min(30000, 2000 * 2 ** attempt++),
      )
  }
  ws.onerror = () => ws.close()
}
export function stopRealtime() {
  active = false
  clearTimeout(timer)
  socket?.close()
  socket = null
  connected.value = false
}
