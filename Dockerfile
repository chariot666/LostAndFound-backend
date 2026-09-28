FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/lost-found-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/admininit ./cmd/admininit

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=build /out/lost-found-server /app/lost-found-server
COPY --from=build /out/admininit /app/admininit

ENV APP_PORT=8080 \
    UPLOAD_DIR=/app/uploads

RUN mkdir -p /app/uploads
EXPOSE 8080

ENTRYPOINT ["/app/lost-found-server"]
