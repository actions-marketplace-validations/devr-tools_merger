FROM golang:1.20 AS build

WORKDIR /src
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/merger-ingest ./cmd/merger-ingest
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/merger-controlplane ./cmd/merger-controlplane

FROM gcr.io/distroless/base-debian11

WORKDIR /app
COPY --from=build /out/merger-ingest /app/merger-ingest
COPY --from=build /out/merger-controlplane /app/merger-controlplane
COPY config /app/config

ENTRYPOINT ["/app/merger-ingest"]
