'use server'
import { createSessionClient } from '@/utils/appwrite/server'
import { ID } from 'node-appwrite'

export async function deleteRealm(realmId: string): Promise<{ error: string | null }> {
    try {
        const { account, databases } = await createSessionClient()
        const user = await account.get()

        const doc = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            realmId
        )
        if (doc.owner_id !== user.$id) {
            return { error: 'Not authorized' }
        }

        await databases.deleteDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            realmId
        )
        return { error: null }
    } catch (e: any) {
        return { error: String(e?.message ?? 'Failed to delete realm') }
    }
}

export async function fetchPlayerCounts(realmIds: string[]): Promise<{ playerCounts: number[] } | null> {
    if (realmIds.length === 0) return null

    try {
        const { account } = await createSessionClient()
        const token = await account.createJWT()

        const url = `${process.env.NEXT_PUBLIC_BACKEND_URL}/api/v1/player-counts?realmIds=${realmIds.join(',')}`
        const res = await fetch(url, { headers: { Authorization: `Bearer ${token.jwt}` } })
        if (!res.ok) return null
        const json = await res.json()
        const payload = json?.data ?? json
        return { playerCounts: payload.playerCounts ?? [] }
    } catch {
        return null
    }
}

export async function createRealm(name: string, mapData: string): Promise<{ data: { $id: string } | null, error: string | null }> {
    try {
        const { account, databases } = await createSessionClient()
        const user = await account.get()

        const document = await databases.createDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            ID.unique(),
            {
                owner_id: user.$id,
                name,
                share_id: ID.unique(),
                only_owner: false,
                map_data: mapData,
            }
        )
        return { data: { $id: String(document.$id) }, error: null }
    } catch (e: any) {
        return { data: null, error: String(e?.message ?? 'Failed to create realm') }
    }
}
