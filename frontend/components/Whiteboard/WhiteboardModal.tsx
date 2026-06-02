'use client'
import React, { useState, useEffect, useRef } from 'react'
import dynamic from 'next/dynamic'
import { server } from '@/utils/backend/server'

// Dynamic import with ssr: false because Excalidraw depends on window/canvas
const Excalidraw = dynamic(
    () => import('@excalidraw/excalidraw').then((mod) => mod.Excalidraw),
    { ssr: false }
)

type WhiteboardModalProps = {
    isOpen: boolean
    setIsOpen: (open: boolean) => void
}

export const WhiteboardModal: React.FC<WhiteboardModalProps> = ({ isOpen, setIsOpen }) => {
    const [excalidrawAPI, setExcalidrawAPI] = useState<any>(null)
    const isUpdatingFromServer = useRef(false)

    useEffect(() => {
        const handleBoardUpdate = (payload: any) => {
            if (!excalidrawAPI || !payload.elements) return
            isUpdatingFromServer.current = true
            excalidrawAPI.updateScene({ elements: payload.elements })
            // Reset the flag shortly after
            setTimeout(() => { isUpdatingFromServer.current = false }, 100)
        }
        server.socket.on('boardUpdate', handleBoardUpdate)
        return () => server.socket.off('boardUpdate', handleBoardUpdate)
    }, [excalidrawAPI])

    const onChange = (elements: readonly any[], appState: any, files: any) => {
        if (isUpdatingFromServer.current) return
        
        // Emit elements to sync with others in the realm
        server.socket.emit('boardUpdate', { elements })
    }

    if (!isOpen) return null

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm p-8">
            <div className="relative w-full h-full bg-white rounded-xl shadow-2xl overflow-hidden flex flex-col">
                <div className="flex justify-between items-center bg-gray-100 p-4 border-b border-gray-300">
                    <h2 className="text-xl font-bold text-gray-800">Collaborative Board</h2>
                    <button 
                        onClick={() => setIsOpen(false)}
                        className="text-gray-500 hover:text-gray-800 transition-colors font-bold text-xl"
                    >
                        ✕
                    </button>
                </div>
                <div className="flex-1 relative w-full h-full overflow-hidden min-h-0">
                    <Excalidraw
                        excalidrawAPI={(api: any) => setExcalidrawAPI(api)}
                        onChange={onChange}
                    />
                </div>
            </div>
        </div>
    )
}
