'use client'
import React, { useState, useEffect, useRef } from 'react'
import { server } from '@/utils/backend/server'

export type ChatMessage = {
    uid: string
    type: 'global' | 'local' | 'room' | 'proximity' | 'direct'
    message: string
}

type ChatPanelProps = {
    isOpen: boolean
    setIsOpen: (open: boolean) => void
    myUid: string
    currentContext: 'room' | 'proximity' | null
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ isOpen, setIsOpen, myUid, currentContext }) => {
    const [messages, setMessages] = useState<ChatMessage[]>([])
    const [activeTab, setActiveTab] = useState<'local' | 'direct'>('local')
    const [targetUid, setTargetUid] = useState<string>('')
    const [input, setInput] = useState('')
    const messagesEndRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        const handleReceive = (msg: ChatMessage) => {
            // We ignore global messages here since they are handled elsewhere or we don't want them polluting the local chat
            if (msg.type !== 'global') {
                setMessages((prev) => [...prev, msg])
            }
        }
        server.socket.on('receiveMessage', handleReceive)
        return () => server.socket.off('receiveMessage', handleReceive)
    }, [])

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages, activeTab, isOpen])

    const handleSend = (e: React.FormEvent) => {
        e.preventDefault()
        if (!input.trim()) return

        // If we are in 'local' tab, we send as 'room' if in a room, or 'proximity' if in proximity
        // Actually, the server routing uses payload type.
        let msgType = 'direct'
        if (activeTab === 'local') {
            msgType = currentContext || 'proximity' // fallback to proximity which will do nothing if no proximity
        }

        const payload: any = { type: msgType, message: input.trim() }
        if (activeTab === 'direct') {
            if (!targetUid.trim()) return
            payload.targetId = targetUid.trim()
        }

        server.socket.emit('sendMessage', payload)
        
        // Optimistically add our own message
        setMessages((prev) => [...prev, { uid: myUid, type: msgType as any, message: input.trim() }])
        setInput('')
    }

    // Determine context label
    let contextLabel = 'Local Chat'
    if (currentContext === 'room') contextLabel = 'Room Chat'
    if (currentContext === 'proximity') contextLabel = 'Proximity Chat'

    const visibleMessages = messages.filter(m => {
        if (activeTab === 'direct') return m.type === 'direct'
        // For local tab, show room, proximity, and local (legacy) messages
        return m.type === 'room' || m.type === 'proximity' || m.type === 'local'
    })

    return (
        <div className={`absolute right-0 top-0 h-full w-80 bg-gray-900/95 backdrop-blur border-l border-gray-700 shadow-2xl flex flex-col z-50 text-white transition-transform duration-300 ${isOpen ? 'translate-x-0' : 'translate-x-full'}`}>
            <div className="p-4 border-b border-gray-700 flex justify-between items-center bg-gray-800/80">
                <h2 className="text-lg font-bold">{activeTab === 'local' ? contextLabel : 'Direct Messages'}</h2>
                <button onClick={() => setIsOpen(false)} className="text-gray-400 hover:text-white transition-colors">
                    ✕
                </button>
            </div>

            <div className="flex border-b border-gray-700 bg-gray-800/80">
                {(['local', 'direct'] as const).map(tab => (
                    <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        className={`flex-1 py-2 text-sm text-center capitalize transition-colors ${activeTab === tab ? 'bg-gray-700 font-bold border-b-2 border-blue-500' : 'hover:bg-gray-700/50 text-gray-400'}`}
                    >
                        {tab}
                    </button>
                ))}
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {visibleMessages.length === 0 && (
                    <div className="text-center text-gray-500 text-sm mt-4">
                        {activeTab === 'local' 
                            ? 'No messages yet. Move close to someone or join a room!' 
                            : 'No direct messages.'}
                    </div>
                )}
                {visibleMessages.map((msg, i) => (
                    <div key={i} className={`flex flex-col ${msg.uid === myUid ? 'items-end' : 'items-start'}`}>
                        <span className="text-xs text-gray-500 mb-1">{msg.uid === myUid ? 'You' : msg.uid.slice(0, 6)}</span>
                        <div className={`px-3 py-2 rounded-lg max-w-[85%] break-words text-sm ${msg.uid === myUid ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-100'}`}>
                            {msg.message}
                        </div>
                    </div>
                ))}
                <div ref={messagesEndRef} />
            </div>

            <form onSubmit={handleSend} className="p-4 border-t border-gray-700 bg-gray-800/80 space-y-2">
                {activeTab === 'direct' && (
                    <input
                        type="text"
                        placeholder="Target User ID..."
                        value={targetUid}
                        onChange={(e) => setTargetUid(e.target.value)}
                        className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-1.5 text-sm outline-none focus:border-blue-500 transition-colors"
                    />
                )}
                <div className="flex gap-2">
                    <input
                        type="text"
                        placeholder={currentContext === null && activeTab === 'local' ? "Nobody nearby..." : "Type a message..."}
                        disabled={currentContext === null && activeTab === 'local'}
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        className="flex-1 bg-gray-900 border border-gray-600 rounded px-3 py-2 text-sm outline-none focus:border-blue-500 transition-colors disabled:opacity-50"
                    />
                    <button type="submit" disabled={currentContext === null && activeTab === 'local'} className="bg-blue-600 hover:bg-blue-500 px-4 py-2 rounded font-medium transition-colors text-sm disabled:opacity-50">
                        Send
                    </button>
                </div>
            </form>
        </div>
    )
}
