import { NextResponse } from "next/server";
import { auth } from "./auth";
import * as jose from "jose";

const INTERNAL_JWT_SECRET = process.env.INTERNAL_JWT_SECRET || "change-me-internal-secret";
const secret = new TextEncoder().encode(INTERNAL_JWT_SECRET);

export const proxy = auth(async (req) => {
  const { pathname } = req.nextUrl;

  // 1. API Proxying with Internal JWT
  if (pathname.startsWith("/api/v1")) {
    const backendUrl = process.env.BACKEND_URL || "http://api";
    const targetUrl = new URL(req.url);
    targetUrl.protocol = new URL(backendUrl).protocol;
    targetUrl.host = new URL(backendUrl).host;
    targetUrl.port = new URL(backendUrl).port;

    const requestHeaders = new Headers(req.headers);
    
    // If authenticated, sign an internal JWT for Go
    if (req.auth?.user) {
      const token = await new jose.SignJWT({ sub: req.auth.user.id })
        .setProtectedHeader({ alg: "HS256" })
        .setIssuedAt()
        .setExpirationTime("1m") // Short lived
        .sign(secret);
      
      requestHeaders.set("Authorization", `Bearer ${token}`);
    }

    return NextResponse.rewrite(targetUrl, {
      request: {
        headers: requestHeaders,
      },
    });
  }

  // 2. Global Route Guards
  const protectedRoutes = ["/admin", "/settings", "/archive", "/d/new"];
  const isProtected = protectedRoutes.some(route => pathname.startsWith(route)) || pathname.endsWith("/edit");

  if (isProtected && !req.auth) {
    const url = new URL("/login", req.url);
    url.searchParams.set("from", pathname);
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    "/api/v1/:path*",
    "/admin/:path*",
    "/settings/:path*",
    "/archive/:path*",
    "/d/new",
    "/d/:postId/edit",
  ],
};
