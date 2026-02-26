import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { toSessionUser, isSessionUser } from "@lib/types/guards";

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

          const rawUser = await res.json();
          const user = toSessionUser(rawUser);
          if (!user) return null;

          return user;
        } catch (error) {
          console.error("Auth authorize error:", error);
          return null;
        }
      },
    }),
  ],
  callbacks: {
    jwt({ token, user }) {
      if (user && isSessionUser(user)) {
        token.user = user;
      }
      return token;
    },
    session({ session, token }) {
      if (token.user && isSessionUser(token.user)) {
        session.user = token.user;
      }
      return session;
    },
  },
  pages: {
    signIn: "/login",
  },
});
