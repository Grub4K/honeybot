FROM golang:1.26.4-alpine@sha256:18b460dd17542c2ba43299a633cf6ebfc1115101509531471d7cfce1019af083 AS go


FROM go AS builder

WORKDIR /app
COPY ./go.mod ./go.sum /app/
RUN --mount=type=cache,target=/go/pkg/mod \
    [ "go", "mod", "download", "-x" ]

COPY . /app
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    [ "go", "build", "-tags", "scratch", "-trimpath", "-ldflags=-s", "-o", "honeybot", "." ]


FROM scratch
COPY --link --from=builder /app/honeybot /honeybot

ARG VERSION
LABEL \
    org.opencontainers.image.title="honeybot" \
    org.opencontainers.image.description="Go Discord Bot for setting up a honeypot channel" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.url="https://github.com/Grub4K/honeybot" \
    org.opencontainers.image.source="https://github.com/Grub4K/honeybot" \
    org.opencontainers.image.version="${VERSION}"

CMD [ "/honeybot" ]
