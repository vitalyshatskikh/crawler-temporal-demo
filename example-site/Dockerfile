FROM golang:1.26 AS builder

WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/siteapi ./cmd/siteapi
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/sitejob ./cmd/sitejob
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/migrator ./cmd/migrator
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/seller ./cmd/seller

FROM scratch
COPY --from=builder /bin/siteapi /bin/siteapi
COPY --from=builder /bin/sitejob /bin/sitejob
COPY --from=builder /bin/migrator /bin/migrator
COPY --from=builder /bin/seller /bin/seller
