# Multi-stage build for ksana-web (Vue 3 + Vite)

# 1) Build stage
FROM node:20-alpine AS build
WORKDIR /app

# Copy package manifests first for better cache
COPY ksana-web/package.json ./

# Install deps (no lockfile present)
RUN npm install

# Copy source
COPY ksana-web/ .

# Build with API base set to /api for nginx proxy
ENV VITE_API_BASE_URL=/api
ENV VITE_API_KEY=
RUN npm run build

# 2) Runtime stage using nginx to serve static files and proxy /api
FROM nginx:1.27-alpine

# Copy custom nginx config
COPY ksana-web/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built static assets
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
