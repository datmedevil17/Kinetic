import { request } from './requests'

// ─────────────────────────────────────────────────────────────────────────────
//  Types
// ─────────────────────────────────────────────────────────────────────────────

type ConnectionResponse = {
    success: boolean
    errorMessage: string
}

// Every message on the wire is { event, payload }
interface WSMessage {
    event: string
    payload: unknown
}

// ─────────────────────────────────────────────────────────────────────────────
//  Server class — native WebSocket, same public API as before
// ─────────────────────────────────────────────────────────────────────────────

const BACKEND_URL: string = process.env.NEXT_PUBLIC_BACKEND_URL as string

class Server {
    private ws: WebSocket | null = null
    private handlers: Map<string, ((data: any) => void)[]> = new Map()
    private _connected: boolean = false
    private _jwt: string = ''

    // ── .socket shim ─────────────────────────────────────────────────────────
    // PlayApp.ts uses server.socket.on / server.socket.off / server.socket.emit
    // This proxy keeps those call sites unchanged — zero edits needed in PlayApp.
    public socket = {
        emit: (event: string, payload?: unknown) => this.emit(event, payload),
        on:   (event: string, handler: (data: any) => void) => this.on(event, handler),
        off:  (event: string, handler: (data: any) => void) => this.off(event, handler),
    }

    // ── connect ───────────────────────────────────────────────────────────────
    // Mirrors: socket.io connect + joinRealm emit
    // Auth: browser WebSocket API cannot set custom headers, so the Appwrite
    //       JWT is passed as a query parameter: /ws?token=<jwt>
    //       The Go server validates it via client.SetJWT() and account.Get()
    public async connect(
        realmId: string,
        uid: string,
        shareId: string,
        access_token: string,
    ): Promise<ConnectionResponse> {
        this._jwt = access_token

        const wsUrl =
            BACKEND_URL.replace(/^http/, 'ws').replace(/\/+$/, '') +
            `/ws?token=${encodeURIComponent(access_token)}`

        this.ws = new WebSocket(wsUrl)

        return new Promise((resolve) => {
            let settled = false
            const settle = (result: ConnectionResponse) => {
                if (!settled) { settled = true; resolve(result) }
            }

            this.ws!.onopen = () => {
                this._connected = true
                this.emit('joinRealm', { realmId, shareId })
            }

            this.ws!.onmessage = (event: MessageEvent) => {
                try {
                    const msg: WSMessage = JSON.parse(event.data as string)
                    this.dispatch(msg.event, msg.payload)

                    if (msg.event === 'joinedRealm') {
                        settle({ success: true, errorMessage: '' })
                    }
                    if (msg.event === 'failedToJoinRoom') {
                        const payload = msg.payload as { reason?: string }
                        settle({ success: false, errorMessage: payload?.reason ?? 'Failed to join realm' })
                    }
                } catch (e) {
                    console.error('[WS] Failed to parse message:', e)
                }
            }

            this.ws!.onerror = () => {
                settle({ success: false, errorMessage: 'Could not connect to server.' })
            }

            this.ws!.onclose = () => {
                this._connected = false
                this.dispatch('disconnect', null)
                settle({ success: false, errorMessage: 'Connection closed.' })
            }
        })
    }

    // ── disconnect ────────────────────────────────────────────────────────────
    public disconnect() {
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
        this._connected = false
    }

    // ── emit (client → server) ────────────────────────────────────────────────
    // Wraps every outgoing message in { event, payload } — the Go hub's envelope
    public emit(event: string, payload?: unknown) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return
        this.ws.send(JSON.stringify({ event, payload }))
    }

    // ── on / off ──────────────────────────────────────────────────────────────
    public on(event: string, handler: (data: any) => void) {
        if (!this.handlers.has(event)) {
            this.handlers.set(event, [])
        }
        this.handlers.get(event)!.push(handler)
    }

    public off(event: string, handler: (data: any) => void) {
        const list = this.handlers.get(event) ?? []
        this.handlers.set(
            event,
            list.filter((h) => h !== handler),
        )
    }

    // ── REST helpers ──────────────────────────────────────────────────────────
    public async getPlayersInRoom(roomIndex: number) {
        if (!this._jwt) return { data: null, error: { message: 'No session' } }
        return request('/api/v1/players-in-room', { roomIndex }, this._jwt)
    }

    // ── internal dispatcher ───────────────────────────────────────────────────
    private dispatch(event: string, payload: unknown) {
        const handlers = this.handlers.get(event) ?? []
        for (const handler of handlers) {
            handler(payload)
        }
    }
}

// Singleton — same pattern as the original socket.io version
const server = new Server()
export { server }