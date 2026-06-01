'use server'
import 'server-only'
import { createSessionClient } from './server'

export async function updateVisitedRealms(shareId: string) {
    const { account, databases } = await createSessionClient()

    let user;
    try {
        user = await account.get()
    } catch (error) {
        return { error }
    }

    try {
        const profile = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
            user.$id
        )

        let visitedRealms = profile.visited_realms || []
        
        // Remove the shareId if it already exists to place it at the front (most recent)
        visitedRealms = visitedRealms.filter((id: string) => id !== shareId)
        
        // Add to the beginning of the array
        visitedRealms.unshift(shareId)

        // Keep only the most recent 10 visited realms
        if (visitedRealms.length > 10) {
            visitedRealms = visitedRealms.slice(0, 10)
        }

        await databases.updateDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
            user.$id,
            { visited_realms: visitedRealms }
        )

        return { error: null }
    } catch (error) {
        return { error }
    }
}
