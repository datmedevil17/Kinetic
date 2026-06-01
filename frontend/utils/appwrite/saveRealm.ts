'use server'
import 'server-only'
import { createSessionClient } from './server'
import { RealmData } from '../pixi/types'
import { RealmDataSchema } from '../pixi/zod'
import { formatForComparison, removeExtraSpaces } from '../removeExtraSpaces'

export async function saveRealm(realmData: RealmData, id: string) {
    const result = RealmDataSchema.safeParse(realmData)
    if (result.success === false) {
        return { error: { message: 'Invalid realm data.' } }
    }

    if (realmData.rooms.length === 0) {
        return { error: { message: 'A realm must have at least one room.' } }
    }

    if (realmData.rooms.length > 50) {
        return { error: { message: 'A realm cannot have more than 50 rooms.' } }
    }

    // return if any rooms in realm data have the same name
    const roomNames = new Set<string>()
    for (const room of realmData.rooms) {
        if (Object.keys(room.tilemap).length > 10_000) {
            return { error: { message: 'This room is too big to save!' } }
        }

        const roomName = formatForComparison(room.name)

        if (roomNames.has(roomName)) {
            return { error: { message: 'Room names must be unique.' } }
        }
        if (roomName.trim() === '') {
            return { error: { message: 'Room name cannot be empty.' } }
        }
        if (roomName.length > 32) {
            return { error: { message: 'Room names cannot be longer than 32 characters.' } }
        }
        roomNames.add(roomName)

        room.name = removeExtraSpaces(room.name, true)
    }

    const { account, databases } = await createSessionClient()

    let user;
    try {
        user = await account.get()
    } catch (error) {
        return { error }
    }

    try {
        // Verify owner_id matches user.$id by fetching the document first
        const doc = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            id
        )

        if (doc.owner_id !== user.$id) {
            return { error: { message: 'Unauthorized' } }
        }

        await databases.updateDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            id,
            {
                map_data: JSON.stringify(realmData)
            }
        )
        return { error: null }
    } catch (error) {
        return { error }
    }
}
