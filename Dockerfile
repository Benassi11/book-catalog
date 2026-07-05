FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=build /server ./server
COPY --from=build /app/web/static ./web/static

EXPOSE 8080

CMD ["./server"]