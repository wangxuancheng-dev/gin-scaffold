# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/server /app/server
COPY configs /app/configs
COPY i18n /app/i18n
ENV TZ=Asia/Shanghai
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
