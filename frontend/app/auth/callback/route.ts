import { Client, Account } from "node-appwrite";
import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function GET(request: Request) {
    const url = new URL(request.url);
    const userId = url.searchParams.get("userId");
    const secret = url.searchParams.get("secret");

    if (!userId || !secret) {
        return NextResponse.redirect(new URL("/signin", request.url));
    }

    try {
        const client = new Client()
            .setEndpoint(process.env.NEXT_PUBLIC_APPWRITE_ENDPOINT!)
            .setProject(process.env.NEXT_PUBLIC_APPWRITE_PROJECT_ID!);

        const account = new Account(client);
        const session = await account.createSession(userId, secret);

        cookies().set("appwrite-session", session.secret, {
            path: "/",
            httpOnly: true,
            sameSite: "strict",
            secure: process.env.NODE_ENV === "production",
        });

        return NextResponse.redirect(new URL("/app", request.url));
    } catch (e) {
        console.error("[auth/callback]", e);
        return NextResponse.redirect(new URL("/signin", request.url));
    }
}
