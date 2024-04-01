# Build stage
FROM golang:alpine AS build

LABEL maintainer="Nathan Moritz"

RUN apk --no-cache add git
# Install dependencies
RUN apk update && apk add --no-cache git && apk add --no-cach bash && apk add build-base

# Set up working directory
RUN mkdir /app
WORKDIR /app

# Copy source code and .env file
COPY . .
COPY .env .

# Download dependencies and build the application
RUN go get -d -v .
RUN go build -o api && cp api /tmp/

# Final stage
FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app/server/
COPY --from=build /tmp/api /app/server/

EXPOSE 8782

CMD ["./api"]