'use client'
import React, { useState, useEffect, useRef } from 'react'
import { server } from '@/utils/backend/server'

type GlobalChatBarProps = {
    isOpen: boolean
    setIsOpen: (open: boolean) => void
    myUid: string
}

export const GlobalChatBar: React.FC<GlobalChatBarProps> = ({ isOpen, setIsOpen, myUid }) => {
    const [input, setInput] = useState('')
    const inputRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        if (isOpen) {
            inputRef.current?.focus()
        }
    }, [isOpen])

    const handleSend = (e: React.FormEvent) => {
        e.preventDefault()
        if (!input.trim()) {
            setIsOpen(false)
            return
        }

        server.socket.emit('sendMessage', { type: 'global', message: input.trim() })
        setInput('')
        setIsOpen(false)
    }

    if (!isOpen) return null

    return (
        <div className="absolute bottom-24 left-1/2 -translate-x-1/2 w-full max-w-2xl z-[60] px-4">
            <form onSubmit={handleSend} className="flex bg-gray-900/95 backdrop-blur border border-gray-700 shadow-2xl rounded-full p-2 items-center">
                <span className="text-blue-400 font-bold px-4 text-sm tracking-widest">GLOBAL</span>
                <input
                    ref={inputRef}
                    type="text"
                    placeholder="Type to everyone (Press Enter to send, Esc to cancel)..."
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === 'Escape') {
                            setIsOpen(false)
                        }
                    }}
                    onBlur={() => {
                        // Small timeout to allow form submission to happen first if clicking a theoretical send button
                        setTimeout(() => setIsOpen(false), 100)
                    }}
                    className="flex-1 bg-transparent text-white px-2 py-2 outline-none text-md"
                />
            </form>
        </div>
    )
}
