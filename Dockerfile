# Build stage
FROM golang:alpine AS build

LABEL maintainer="Nathan Moritz"

# Install dependencies
RUN apk update && apk add --no-cache git bash build-base

# Set up working directory
WORKDIR /app

# Copy source code and .env file
COPY . .
COPY .env .

# Download dependencies and build the application
RUN go get -d -v .
RUN go build -o api

# Final stage
FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app/server/

# Copy the built binary from the build stage
COPY --from=build /app/api /app/server/

EXPOSE 8782

CMD ["./api"]