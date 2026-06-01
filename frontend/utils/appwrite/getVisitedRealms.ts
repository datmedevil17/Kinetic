'use server'
import 'server-only'
import { createSessionClient } from './server'
import { Query } from 'node-appwrite'

export async function getVisitedRealms() {
    const { account, databases } = await createSessionClient()

    let user;
    try {
        user = await account.get()
    } catch (error) {
        return { data: null, error }
    }
    
    let profile;
    try {
        profile = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
            user.$id
        )
    } catch (error) {
        return { data: null, error }
    }

    const visitedRealms = []
    const realmsToRemove: string[] = []
    for (const shareId of profile.visited_realms) {
        try {
            const data = await databases.listDocuments(
                process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
                process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
                [Query.equal('share_id', shareId)]
            )
            
            if (data.documents.length > 0) {
                visitedRealms.push(data.documents[0])
            } else {
                realmsToRemove.push(shareId)
            }
        } catch {
            realmsToRemove.push(shareId)
        }
    }

    if (realmsToRemove.length > 0) {
        const newVisited = profile.visited_realms.filter((shareId: string) => !realmsToRemove.includes(shareId));
        await databases.updateDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
            user.$id,
            { 
                visited_realms: newVisited
            }
        )
    }

    return { data: visitedRealms, error: null }
}
