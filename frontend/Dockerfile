# frontend/Dockerfile

FROM node:20-alpine AS build
WORKDIR /app

# Copy dependency manifests from the *frontend* context root
COPY package.json ./
# Copy whichever lockfile(s) exist (safe even if some don't)
COPY package-lock.json* pnpm-lock.yaml* yarn.lock* ./

# Install deps
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Copy the rest of the frontend source
COPY . .

# Build static assets
RUN npm run build

FROM nginx:1.27-alpine

# Nginx config is also in the frontend context root
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
