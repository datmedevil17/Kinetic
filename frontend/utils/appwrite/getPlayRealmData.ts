'use server'
import 'server-only'
import { createSessionClient } from './server'

export async function getPlayRealmData(shareId: string) {
    const { account, databases } = await createSessionClient()

    let user;
    try {
        user = await account.get()
    } catch (error) {
        return { data: null, error }
    }

    try {
        const data = await databases.listDocuments(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            [
                // Query.equal('share_id', shareId) // Assuming Query is imported from node-appwrite
                // We'll just construct the string directly to avoid extra imports if we want:
                `equal("share_id", ["${shareId}"])`
            ]
        )

        if (data.documents.length === 0) {
            return { data: null, error: { message: 'Realm not found' } }
        }

        const realm = data.documents[0]

        // if we are the owner, always return the data
        if (realm.owner_id === user.$id) {
            return { data: realm, error: null }
        }

        if (realm.only_owner) {
            return { data: null, error: { message: 'only owner' }}
        }

        return { data: realm, error: null }
    } catch (error) {
        return { data: null, error }
    }
}
