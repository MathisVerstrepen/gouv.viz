FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN templ generate -path ./web/components
RUN CGO_ENABLED=0 GOOS=linux go build -o gouv-viz ./cmd/web

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/gouv-viz ./gouv-viz
COPY --from=builder /app/web/assets ./web/assets

ENV ENV=prod
ENV PORT=9456
ENV ASSETS_PATH=web/assets
ENV DATABASE_PATH=/data/gouv-viz.sqlite

RUN mkdir -p /data
VOLUME ["/data"]
EXPOSE 9456

CMD ["./gouv-viz"]
