# 1) BUILD API
FROM golang:alpine AS build-go
# Add Maintainer info
LABEL maintainer="Nathan Moritz"
# Install git.
RUN apk --no-cache add git
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git && apk add --no-cach bash && apk add build-base
# Setup folders
RUN mkdir /app
WORKDIR /app
# Copy the source from the current directory to the working Directory inside the container
COPY . .
COPY .env .
# Download all the dependencies
RUN go get -d -v .
# build
RUN go build -o api && cp api /tmp/

# 2) BUILD FINAL IMAGE
FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app/server/
COPY --from=build-go /tmp/api /app/server/
EXPOSE 8782
CMD ["./api"]