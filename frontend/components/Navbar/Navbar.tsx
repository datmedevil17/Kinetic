import React from 'react'
import { createSessionClient } from '@/utils/appwrite/server'
import { NavbarChild } from './NavbarChild'
import { formatEmailToName } from '@/utils/formatEmailToName'

export const Navbar:React.FC = async () => {

    const { account } = await createSessionClient()
    let user = null;
    try {
        user = await account.get()
    } catch {
        // user not logged in
    }

    return (
        <NavbarChild name={user ? formatEmailToName(user.email) : undefined} avatar_url={undefined}/>
    )
}
