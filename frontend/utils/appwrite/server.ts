import { Client, Account, Databases } from 'node-appwrite';
import { cookies } from 'next/headers';

export const createSessionClient = async () => {
    const client = new Client()
        .setEndpoint(process.env.NEXT_PUBLIC_APPWRITE_ENDPOINT!)
        .setProject(process.env.NEXT_PUBLIC_APPWRITE_PROJECT_ID!);

    // Extract the fallback session cookie if it exists
    const sessionCookie = cookies().get('appwrite-session');
    
    if (sessionCookie && sessionCookie.value) {
        client.setSession(sessionCookie.value);
    }

    return {
        client,
        account: new Account(client),
        databases: new Databases(client)
    };
};

export const createAdminClient = () => {
    const client = new Client()
        .setEndpoint(process.env.NEXT_PUBLIC_APPWRITE_ENDPOINT!)
        .setProject(process.env.NEXT_PUBLIC_APPWRITE_PROJECT_ID!)
        // Needs a server-side API key
        .setKey(process.env.APPWRITE_API_KEY!);

    return {
        client,
        account: new Account(client),
        databases: new Databases(client)
    };
};
