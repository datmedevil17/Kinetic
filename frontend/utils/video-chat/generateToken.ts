'use server'
import { RtcRole, RtcTokenBuilder } from 'agora-token'
import { createSessionClient } from '../appwrite/server'

export async function generateToken(channelName: string) {
    const { account } = await createSessionClient()
    try {
        await account.get()
    } catch {
        return null
    }

    const appId = process.env.NEXT_PUBLIC_AGORA_APP_ID
    const appCertificate = process.env.APP_CERTIFICATE
    if (!appId || !appCertificate) return null
    const uid = 0
    const role = RtcRole.PUBLISHER
    const expireTime = 3600
    const currentTimestamp = Math.floor(Date.now() / 1000)
    const expiredTs = currentTimestamp + expireTime

    const token = RtcTokenBuilder.buildTokenWithUid(
        appId,
        appCertificate,
        channelName,
        uid,
        role,
        expiredTs,
        expiredTs,
    )

    return token
}