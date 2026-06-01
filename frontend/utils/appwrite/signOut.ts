'use server'
import { cookies } from 'next/headers'
import { createSessionClient } from './server'

export async function signOut() {
    try {
        const { account } = await createSessionClient()
        await account.deleteSession('current')
    } catch {
        // session may already be invalid
    }
    cookies().delete('appwrite-session')
}
