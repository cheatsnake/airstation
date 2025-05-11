FROM node:22-alpine AS player
WORKDIR /app
COPY ./web/player/package*.json ./
RUN npm install
COPY ./web/player .
RUN npm run build

FROM node:22-alpine AS studio
WORKDIR /app
COPY ./web/studio/package*.json ./
RUN npm install
COPY ./web/studio .
RUN npm run build

FROM golang:1.24-alpine AS server
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /app/bin/main ./cmd/main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ffmpeg
COPY --from=server /app/bin/main .
COPY --from=player /app/dist ./player
COPY --from=studio /app/dist ./studio
EXPOSE 7331
ENTRYPOINT ["./main"]