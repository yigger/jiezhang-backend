# syntax=docker/dockerfile:1
FROM golang:1.26.3-bookworm

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/air-verse/air@v1.66.0 \
    && cp /go/bin/air /usr/local/bin/air

COPY docker/dev-entrypoint.sh /usr/local/bin/dev-entrypoint.sh
RUN chmod +x /usr/local/bin/dev-entrypoint.sh

EXPOSE 10240
ENTRYPOINT ["/usr/local/bin/dev-entrypoint.sh"]
CMD ["air", "-c", ".air.toml"]
