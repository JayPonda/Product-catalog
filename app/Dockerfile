# syntax=docker/dockerfile:1

# ---------- Build stage ----------
FROM node:22-alpine AS build

RUN corepack enable && corepack prepare pnpm@10 --activate

WORKDIR /app

# Copy manifests first to cache the dependency layer.
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

# Backend URL is baked into the bundle at build time.
# Override per environment: docker build --build-arg VITE_BACKEND_URL=https://api.example.com/
ARG VITE_BACKEND_URL=http://localhost:3300/
ENV VITE_BACKEND_URL=$VITE_BACKEND_URL

RUN pnpm build

# ---------- Runtime stage ----------
FROM nginx:1.29-alpine

# Drop default config; serve only our SPA with hardened headers.
RUN rm /etc/nginx/conf.d/default.conf
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://127.0.0.1:80/ || exit 1
