import { createAdminClient } from "@/utils/appwrite/server";
import { NextResponse } from "next/server";

export async function GET(request: Request) {
    const url = new URL(request.url);
    const userId = url.searchParams.get("userId");
    const secret = url.searchParams.get("secret");

    if (!userId || !secret) {
        return NextResponse.redirect(new URL("/signin", request.url));
    }

    try {
        const { account, databases } = createAdminClient();
        const session = await account.createSession(userId, secret);

        // Create a profile document for this user if one doesn't exist yet
        try {
            await databases.getDocument(
                process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
                process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
                userId
            );
        } catch {
            // Profile doesn't exist — create it
            await databases.createDocument(
                process.env.NEXT_PUBLIC_APPWRITE_DATABASE_ID!,
                process.env.NEXT_PUBLIC_APPWRITE_PROFILES_COLLECTION_ID!,
                userId,
                { skin: "009", visited_realms: [] }
            );
        }

        const secure = process.env.NODE_ENV === "production";
        const cookie = `appwrite-session=${session.secret}; Path=/; HttpOnly; SameSite=Lax${secure ? "; Secure" : ""}`;

        const response = NextResponse.redirect(new URL("/app", request.url));
        response.headers.append("Set-Cookie", cookie);
        return response;
    } catch (e) {
        console.error("[auth/callback]", e);
        return NextResponse.redirect(new URL("/signin", request.url));
    }
}
