'use client'
import React, { useEffect, useState } from 'react'
import PixiApp from './PixiApp'
import { RealmData } from '@/utils/pixi/types'
import PlayNavbar from './PlayNavbar'
import { useModal } from '../hooks/useModal'
import signal from '@/utils/signal'
import IntroScreen from './IntroScreen'
import VideoBar from '@/components/VideoChat/VideoBar'
import { AgoraVideoChatProvider } from '../hooks/useVideoChat'
import { ChatPanel } from '@/components/Chat/ChatPanel'
import { GlobalChatBar } from '@/components/Chat/GlobalChatBar'
import { WhiteboardModal } from '@/components/Whiteboard/WhiteboardModal'
import { toast } from 'react-toastify'
import { server } from '@/utils/backend/server'

type PlayClientProps = {
    mapData: RealmData
    username: string
    access_token: string
    realmId: string
    uid: string
    shareId: string
    initialSkin: string
    name: string
}

const PlayClient:React.FC<PlayClientProps> = ({ mapData, username, access_token, realmId, uid, shareId, initialSkin, name }) => {

    const { setErrorModal, setDisconnectedMessage } = useModal()

    const [showIntroScreen, setShowIntroScreen] = useState(true)

    const [skin, setSkin] = useState(initialSkin)
    
    const [isChatOpen, setIsChatOpen] = useState(false)
    const [isWhiteboardOpen, setIsWhiteboardOpen] = useState(false)
    const [isGlobalChatBarOpen, setIsGlobalChatBarOpen] = useState(false)
    const [currentContext, setCurrentContext] = useState<'room' | 'proximity' | null>(null)

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            // Ignore shortcut if typing in an input
            if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return

            if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'e') {
                e.preventDefault()
                setIsWhiteboardOpen(prev => !prev)
            }
            if ((e.metaKey || e.ctrlKey) && e.key === '/') {
                e.preventDefault()
                setIsGlobalChatBarOpen(prev => !prev)
            }
        }
        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [])

    useEffect(() => {
        const onShowKickedModal = (message: string) => { 
            setErrorModal('Disconnected')
            setDisconnectedMessage(message)
        }

        const onShowDisconnectModal = () => {
            setErrorModal('Disconnected')
            setDisconnectedMessage('You have been disconnected from the server.')
        }

        const onSwitchSkin = (skin: string) => {
            setSkin(skin)
        }

        const onChatContextUpdate = (context: 'room' | 'proximity' | null) => {
            setCurrentContext(context)
            if (context !== null) {
                setIsChatOpen(true)
            }
        }

        const onReceiveMessage = (msg: any) => {
            if (msg.type === 'global') {
                toast.info(`🌍 ${msg.uid === uid ? 'You' : msg.uid.slice(0, 6)}: ${msg.message}`, {
                    position: "top-center",
                    autoClose: 6000,
                    hideProgressBar: true,
                    closeOnClick: true,
                    pauseOnHover: true,
                    draggable: true,
                    theme: "dark",
                    style: { borderRadius: '9999px', fontSize: '14px', border: '1px solid #374151', backgroundColor: '#111827' }
                })
            }
        }

        signal.on('showKickedModal', onShowKickedModal)
        signal.on('showDisconnectModal', onShowDisconnectModal)
        signal.on('switchSkin', onSwitchSkin)
        signal.on('chatContextUpdate', onChatContextUpdate)
        server.socket.on('receiveMessage', onReceiveMessage)

        return () => {
            signal.off('showKickedModal', onShowDisconnectModal)
            signal.off('showDisconnectModal', onShowDisconnectModal)
            signal.off('switchSkin', onSwitchSkin)
            signal.off('chatContextUpdate', onChatContextUpdate)
            server.socket.off('receiveMessage', onReceiveMessage)
        }
    }, [])

    return (
        <AgoraVideoChatProvider uid={uid}>
            {!showIntroScreen && <div className='relative w-full h-screen flex flex-col-reverse sm:flex-col overflow-hidden'>
                <VideoBar />
                <PixiApp 
                    mapData={mapData} 
                    className='w-full grow sm:h-full sm:flex-grow-0' 
                    username={username} 
                    access_token={access_token} 
                    realmId={realmId} 
                    uid={uid} 
                    shareId={shareId} 
                    initialSkin={skin} 
                />
                <PlayNavbar username={username} skin={skin}/>
                
                <GlobalChatBar isOpen={isGlobalChatBarOpen} setIsOpen={setIsGlobalChatBarOpen} myUid={uid} />
                <ChatPanel isOpen={isChatOpen} setIsOpen={setIsChatOpen} myUid={uid} currentContext={currentContext} />
                <WhiteboardModal isOpen={isWhiteboardOpen} setIsOpen={setIsWhiteboardOpen} />
                
                <div className="absolute bottom-4 right-4 flex flex-col gap-2 z-40">
                    <button onClick={() => setIsWhiteboardOpen(true)} className="bg-gray-800 text-white p-3 rounded-full shadow-lg hover:bg-gray-700 transition" title="Whiteboard (Cmd+E)">
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                    </button>
                    <button onClick={() => setIsChatOpen(true)} className="bg-blue-600 text-white p-3 rounded-full shadow-lg hover:bg-blue-500 transition" title="Local Chat">
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
                    </button>
                </div>
            </div>}
            {showIntroScreen && <IntroScreen realmName={name} skin={skin} username={username} setShowIntroScreen={setShowIntroScreen}/>}    
        </AgoraVideoChatProvider>
    )
}
export default PlayClient