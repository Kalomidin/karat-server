FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/karat-server ./cmd/karat-server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/karat-server /usr/local/bin/karat-server
USER nobody
EXPOSE 8080
ENTRYPOINT ["karat-server"]
