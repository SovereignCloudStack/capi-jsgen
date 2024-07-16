############# builder
FROM golang:1.22-alpine AS builder
ARG VERSION
ENV BINARY_PATH=/go/bin
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w -X 'main.version=$VERSION'" -o /capi-jsgen

############# capi-jsgen
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/SovereignCloudStack/capi-jsgen"

COPY --from=builder /capi-jsgen /
# could be removed, just for dev-purposes
COPY --from=builder /app/data /data
EXPOSE 8080
ENTRYPOINT ["/capi-jsgen"]