type Handler = (msg: any) => void

export class ManagerHubWS {
  private ws: WebSocket | null = null
  private handlers = new Map<string, Set<Handler>>()
  private queue: string[] = []

  constructor(private token: string) {}

  connect() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/api/v1/ws?token=${this.token}`
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      for (const m of this.queue) this.ws?.send(m)
      this.queue = []
    }
    this.ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        const event = msg.event || msg.type
        const data = msg.data ?? msg.payload ?? msg
        for (const h of this.handlers.get(event) || []) h(data)
        for (const h of this.handlers.get('*') || []) h({ event, data })
      } catch {}
    }
    this.ws.onclose = () => {
      setTimeout(() => this.connect(), 3000)
    }
  }

  send(obj: any) {
    const raw = JSON.stringify(obj)
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(raw)
    } else {
      this.queue.push(raw)
    }
  }

  on(event: string, h: Handler): () => void {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set())
    this.handlers.get(event)!.add(h)
    return () => { this.handlers.get(event)?.delete(h) }
  }

  close() {
    this.ws?.close()
    this.ws = null
  }
}
