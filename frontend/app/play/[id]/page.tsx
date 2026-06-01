import React from 'react'
import NotFound from '@/app/not-found'
import { createSessionClient, createAdminClient } from '@/utils/appwrite/server'
import { redirect } from 'next/navigation'
import { getPlayRealmData } from '@/utils/appwrite/getPlayRealmData'
import PlayClient from '../PlayClient'
import { updateVisitedRealms } from '@/utils/appwrite/updateVisitedRealms'
import { formatEmailToName } from '@/utils/formatEmailToName'

export default async function Play({ params, searchParams }: { params: { id: string }, searchParams: { shareId: string } }) {

    const { account, databases } = await createSessionClient()
    let user;
    try {
        user = await account.get()
    } catch {
        return redirect('/signin')
    }
    
    let jwt;
    try {
        const session = await account.createJWT()
        jwt = session.jwt
    } catch {
        return redirect('/signin')
    }

    let data, profile, error: any, profileError: any;
    try {
        if (!searchParams.shareId) {
            data = await databases.getDocument(
                process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
                process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
                params.id
            )
        } else {
            const res = await getPlayRealmData(searchParams.shareId)
            data = res.data
            error = res.error
        }
    } catch (e: any) {
        error = e
    }

    try {
        profile = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
            user.$id
        )
    } catch {
        // Profile doesn't exist yet — create it with admin client
        try {
            const { databases: adminDb } = createAdminClient()
            profile = await adminDb.createDocument(
                process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
                process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
                user.$id,
                { skin: '009', visited_realms: [] }
            )
        } catch (e: any) {
            profileError = e
        }
    }

    if (!data || !profile) {
        const message = error?.message || profileError?.message
        console.error('[play] data:', !!data, 'profile:', !!profile, 'error:', message)
        return <NotFound specialMessage={message}/>
    }

    const realm = data
    const map_data = typeof realm.map_data === 'string' ? JSON.parse(realm.map_data) : realm.map_data

    let skin = profile.skin && profile.skin !== 'default' ? profile.skin : '009'

    if (searchParams.shareId && realm.owner_id !== user.$id) {
        updateVisitedRealms(searchParams.shareId)
    }

    return (
        <PlayClient 
            mapData={map_data} 
            username={formatEmailToName(user.email)} 
            access_token={jwt} 
            realmId={params.id} 
            uid={user.$id} 
            shareId={searchParams.shareId || ''} 
            initialSkin={skin}
            name={realm.name}
        />
    )
}