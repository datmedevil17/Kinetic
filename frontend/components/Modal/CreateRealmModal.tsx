'use client'
import React, { useState } from 'react'
import Modal from './Modal'
import { useModal } from '@/app/hooks/useModal'
import BasicButton from '../BasicButton'
import BasicInput from '../BasicInput'
import { createClient } from '@/utils/appwrite/client'
import { ID } from 'appwrite'
import { toast } from 'react-toastify'
import { useRouter } from 'next/navigation' 
import revalidate from '@/utils/revalidate'
import { removeExtraSpaces } from '@/utils/removeExtraSpaces'
import defaultMap from '@/utils/defaultmap.json'

const CreateRealmModal:React.FC = () => {
    
    const { modal, setModal } = useModal()
    const [realmName, setRealmName] = useState<string>('')
    const [loading, setLoading] = useState<boolean>(false)

    const [useDefaultMap, setUseDefaultMap] = useState<boolean>(true)

    const router = useRouter()

    async function createRealm() {
        const { account, databases } = createClient()
        let user;
        try {
            user = await account.get()
        } catch {
            return
        }

        const uid = user.$id

        setLoading(true)

        const realmData: any = {
            owner_id: uid,
            name: realmName,
            share_id: ID.unique(), // Appwrite ID generator for share_id
            only_owner: false
        }
        
        if (useDefaultMap) {
            realmData.map_data = JSON.stringify(defaultMap)
        } else {
            realmData.map_data = JSON.stringify({ rooms: [] })
        }

        try {
            const document = await databases.createDocument(
                process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
                process.env.NEXT_PUBLIC_APPWRITE_REALMS_COLLECTION_ID!,
                ID.unique(),
                realmData
            )
            
            setRealmName('')
            revalidate('/app')
            setModal('None')
            toast.success('Your space has been created!')
            router.push(`/editor/${document.$id}`)
        } catch (error: any) {
            toast.error(error?.message)
        }

        setLoading(false)
    }

    function onChange(e: React.ChangeEvent<HTMLInputElement>) {
        const value = removeExtraSpaces(e.target.value)
        setRealmName(value)
    }

    return (
        <Modal open={modal === 'Create Realm'} closeOnOutsideClick>
            <div className='flex flex-col items-center p-4 w-[400px] gap-4'>
                <h1 className='text-2xl'>Create a Space</h1>
                <BasicInput label={'Space Name'} className='w-[280px]' value={realmName} onChange={onChange} maxLength={32}/>
                <div className='flex items-center gap-2 w-[280px]'>
                    <input
                        type="checkbox"
                        id="useDefaultMap"
                        checked={useDefaultMap}
                        onChange={(e) => setUseDefaultMap(e.target.checked)}
                    />
                    <label htmlFor="useDefaultMap">Use starter map</label>
                </div>
                <BasicButton disabled={realmName.length <= 0 || loading} onClick={createRealm} className='text-lg'>
                    Create
                </BasicButton>
            </div>
        </Modal>
    )
}

export default CreateRealmModal