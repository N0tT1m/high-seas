# Build stage
FROM node:18.18-alpine3.18 AS build

LABEL maintainer="Nathan Moritz"

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

RUN npm run build

# Final stage
FROM nginx:latest

COPY --from=build /app/dist/high-seas-frontend /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
COPY ./fullchain.pem /etc/letsencrypt/live/goose.cinemacloud.tv/fullchain.pem
COPY ./privkey.pem /etc/letsencrypt/live/goose.cinemacloud.tv/privkey.pem 

EXPOSE 80 443
