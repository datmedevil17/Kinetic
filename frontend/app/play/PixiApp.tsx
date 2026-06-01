'use client'
import React, { useRef } from 'react'
import { PlayApp } from '@/utils/pixi/PlayApp'
import { useEffect } from 'react'
import { RealmData } from '@/utils/pixi/types'
import { useModal } from '../hooks/useModal'
import { server } from '@/utils/backend/server'

type PixiAppProps = {
    className?: string
    mapData: RealmData
    username: string
    access_token: string
    realmId: string
    uid: string
    shareId: string
    initialSkin: string
}

const PixiApp:React.FC<PixiAppProps> = ({ className, mapData, username, access_token, realmId, uid, shareId, initialSkin }) => {

    const appRef = useRef<PlayApp | null>(null)
    const { setModal, setLoadingText, setFailedConnectionMessage, setErrorModal } = useModal()

    useEffect(() => {
        let mounted = true

        const mount = async () => {
            setModal('Loading')
            setLoadingText('Connecting to server...')

            const { success, errorMessage } = await server.connect(realmId, uid, shareId, access_token)

            if (!mounted) return  // cleanup ran while we were connecting

            if (!success) {
                setErrorModal('Failed To Connect')
                setFailedConnectionMessage(errorMessage)
                return
            }

            const app = new PlayApp(uid, realmId, mapData, username, initialSkin)
            appRef.current = app

            setLoadingText('Loading game...')
            await app.init()

            if (!mounted) return  // cleanup ran while we were loading

            setModal('None')
            document.getElementById('app-container')!.appendChild(app.getApp().canvas)
        }

        mount()

        return () => {
            mounted = false
            if (appRef.current) {
                appRef.current.destroy()
                appRef.current = null
            } else {
                server.disconnect()
            }
        }
    }, [])

    return (
        <div id='app-container' className={`overflow-hidden ${className}`}>
            
        </div>
    )
}

export default PixiApp
