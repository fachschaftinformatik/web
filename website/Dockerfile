FROM node:24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY website/package.json website/package-lock.json ./website/
RUN --mount=type=cache,target=/root/.npm cd website && npm ci
COPY website/ ./website
COPY docs/ ./docs
RUN --mount=type=cache,target=/app/website/node_modules/.vite cd website && npm run build

FROM nginx:alpine AS production
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/website/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
