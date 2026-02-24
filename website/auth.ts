import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import type { DtoUserResponse as User } from "@lib/api";

export const { handlers, auth, signIn, signOut } = NextAuth({
  trustHost: true,
  basePath: "/api/auth",
  providers: [
    Credentials({
      name: "Credentials",
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) return null;

        try {
          const backendUrl = process.env.BACKEND_URL || "http://api";
          const res = await fetch(`${backendUrl}/api/v1/auth/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              email: credentials.email,
              password: credentials.password,
            }),
          });

          if (!res.ok) return null;

          const user = await res.json() as User;
          return {
            id: String(user.id),
            name: user.name,
            email: user.email,
            role: user.role,
            avatar_url: user.avatar_url,
            program_id: user.program_id,
            verified: user.verified,
            active: user.active,
            theme: user.theme,
            private: user.private,
          };
        } catch (error) {
          console.error("Auth authorize error:", error);
          return null;
        }
      },
    }),
  ],
  callbacks: {
    jwt({ token, user }) {
      if (user) {
        token.user = user;
      }
      return token;
    },
    session({ session, token }) {
      if (token.user) {
        session.user = token.user as any;
      }
      return session;
    },
  },
  pages: {
    signIn: "/login",
  },
});
