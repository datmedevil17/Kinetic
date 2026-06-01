import { createSessionClient } from '@/utils/appwrite/server'
import { redirect } from 'next/navigation'
import ManageChild from '../ManageChild'
import NotFound from '../../not-found'
import { request } from '@/utils/backend/requests'

export default async function Manage({ params }: { params: { id: string } }) {

    const { account, databases } = await createSessionClient()

    let user;
    try {
        user = await account.get()
    } catch {
        return redirect('/signin')
    }

    try {
        const data = await databases.getDocument(
            process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
            process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
            params.id
        )
        const realm = data

    return (
        <div>
            <ManageChild 
                realmId={realm.id} 
                startingShareId={realm.share_id} 
                startingOnlyOwner={realm.only_owner} 
                startingName={realm.name}
            />
        </div>
    )
}