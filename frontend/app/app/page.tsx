import { redirect } from 'next/navigation'
import { Navbar } from '@/components/Navbar/Navbar'
import RealmsMenu from './RealmsMenu/RealmsMenu'
import { createSessionClient } from '@/utils/appwrite/server'
import { getVisitedRealms } from '@/utils/appwrite/getVisitedRealms'
import { Query } from 'node-appwrite'

export default async function App() {

    const { account, databases } = await createSessionClient()

    let user;
    try {
        user = await account.get()
    } catch {
        return redirect('/signin')
    }

    const realms: any = []
    const { data: ownedRealms, error } = await databases.listDocuments(
        process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
        process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
        [Query.equal('owner_id', user.$id)]
    ).then(res => ({ data: res.documents, error: null })).catch(e => ({ data: null, error: e }))

    if (ownedRealms) {
        realms.push(...ownedRealms)
    }

    const { data: visitedRealms } = await getVisitedRealms()
    if (visitedRealms) {
        const tagged = visitedRealms.map((realm: any) => ({ ...realm, shared: true }))
        realms.push(...tagged)
    }

    const errorMessage = error?.message || ''

    return (
        <div className='w-full min-h-screen bg-gray-900 relative overflow-hidden'>
            {/* Background Elements */}
            <div className='absolute inset-0 bg-gray-800/50'></div>
            <div className='absolute top-20 left-20 w-72 h-72 bg-gray-700/20 rounded-full blur-3xl'></div>
            <div className='absolute bottom-20 right-20 w-96 h-96 bg-gray-600/20 rounded-full blur-3xl'></div>

            <div className='relative z-10'>
                <Navbar />
                <div className='px-4 sm:px-8 pt-8 pb-8'>
                    <h1 className='text-4xl font-bold text-white mb-2'>Your Spaces</h1>
                    <p className='text-gray-300 mb-8'>Manage and access your virtual collaboration spaces</p>
                    <RealmsMenu realms={realms} errorMessage={errorMessage} />
                </div>
            </div>
        </div>
    )
}
