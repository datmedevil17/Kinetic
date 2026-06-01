import NotFound from '@/app/not-found'
import { createSessionClient } from '@/utils/appwrite/server'
import { redirect } from 'next/navigation'
import Editor from '../Editor'

export default async function RealmEditor({ params }: { params: { id: string } }) {

    const { account, databases } = await createSessionClient()

    let user: any
    try {
        user = await account.get()
    } catch {
        return redirect('/signin')
    }

    let data: any
    try {
        data = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            params.id
        )
    } catch {
        return <NotFound />
    }

    if (data.owner_id !== user.$id) {
        return <NotFound />
    }

    const map_data = data.map_data

    return (
        <div>
            <Editor realmData={map_data}/>
        </div>
    )
}
