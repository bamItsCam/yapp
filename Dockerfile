FROM golang:1.21-bookworm as builder

WORKDIR /build

# Copy local code to the container image.
COPY . ./

RUN CGO_ENABLED=0 go build -v -o app

# Application image.
FROM gcr.io/distroless/base:latest

COPY --from=builder /build/app /usr/local/bin/app

EXPOSE 3000

CMD ["/usr/local/bin/app"]