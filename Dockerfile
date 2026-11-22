FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/karat-server ./cmd/karat-server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/karat-server /app/karat-server
COPY configs /app/configs
USER nobody
EXPOSE 8080
ENTRYPOINT ["/app/karat-server"]
