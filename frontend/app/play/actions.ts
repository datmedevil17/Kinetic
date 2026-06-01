'use server'
import { createSessionClient } from '@/utils/appwrite/server'

export async function updateSkin(skin: string): Promise<{ error: string | null }> {
    try {
        const { account, databases } = await createSessionClient()
        const user = await account.get()
        await databases.updateDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
            user.$id,
            { skin }
        )
        return { error: null }
    } catch (e: any) {
        return { error: String(e?.message ?? 'Failed to update skin') }
    }
}
