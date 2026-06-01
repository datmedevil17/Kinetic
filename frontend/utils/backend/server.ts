import { createClient } from '../appwrite/client'
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
        // Convert http(s):// → ws(s)://  and append the auth token
        const wsUrl =
            BACKEND_URL.replace(/^http/, 'ws').replace(/\/+$/, '') +
            `/ws?token=${encodeURIComponent(access_token)}`

        this.ws = new WebSocket(wsUrl)

        return new Promise((resolve) => {
            this.ws!.onopen = () => {
                this._connected = true
                // Mirror: this.socket.emit('joinRealm', { realmId, shareId })
                this.emit('joinRealm', { realmId, shareId })
            }

            this.ws!.onmessage = (event: MessageEvent) => {
                try {
                    const msg: WSMessage = JSON.parse(event.data as string)
                    this.dispatch(msg.event, msg.payload)

                    // Resolve the connect() promise once we know the join outcome
                    if (msg.event === 'joinedRealm') {
                        resolve({ success: true, errorMessage: '' })
                    }
                    if (msg.event === 'failedToJoinRoom') {
                        const payload = msg.payload as { reason?: string }
                        resolve({
                            success: false,
                            errorMessage: payload?.reason ?? 'Failed to join realm',
                        })
                    }
                } catch (e) {
                    console.error('[WS] Failed to parse message:', e)
                }
            }

            this.ws!.onerror = () => {
                resolve({
                    success: false,
                    errorMessage: 'Could not connect to server.',
                })
            }

            this.ws!.onclose = () => {
                this._connected = false
                // Fire the 'disconnect' event — PlayApp listens for this
                this.dispatch('disconnect', null)
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
    // Path changed: /getPlayersInRoom → /api/v1/players-in-room
    public async getPlayersInRoom(roomIndex: number) {
        const { account } = createClient()
        let session;
        try {
            session = await account.createJWT()
        } catch {
            return { data: null, error: { message: 'No session provided' } }
        }
        return request('/api/v1/players-in-room', { roomIndex }, session.jwt)
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