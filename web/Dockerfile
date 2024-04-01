# Build stage
FROM node:18.18-alpine3.18 AS build

LABEL maintainer="Nathan Moritz"

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

RUN npm run build-webpack

# Final stage
FROM nginx:latest

COPY --from=build /app/dist/high-seas-frontend /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
