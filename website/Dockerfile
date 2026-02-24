FROM --platform=$BUILDPLATFORM oven/bun:alpine AS base
FROM base AS deps
RUN apk add --no-cache libc6-compat git
WORKDIR /app
COPY website/package.json website/bun.lock ./
RUN bun install --frozen-lockfile

FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY website/ .
COPY docs/ ../docs

RUN bun run build
FROM node:24-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
COPY --from=builder /app/public ./public
RUN mkdir .next
RUN chown nextjs:nodejs .next
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT 80
CMD ["node", "server.js"]
