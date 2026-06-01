import { Client, Account, Databases } from 'appwrite';

export const createClient = () => {
    const client = new Client()
        .setEndpoint(process.env.NEXT_PUBLIC_APPWRITE_ENDPOINT!)
        .setProject(process.env.NEXT_PUBLIC_APPWRITE_PROJECT_ID!);

    return {
        client,
        account: new Account(client),
        databases: new Databases(client)
    };
};
