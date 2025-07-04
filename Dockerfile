FROM cgr.dev/chainguard/busybox:latest

ARG ARCH="amd64"
ARG OS="linux"
COPY bin/hostport-allocator-manager-${OS}-${ARCH} /bin/hostport-allocator-manager
USER nobody

ENTRYPOINT [ "/bin/hostport-allocator-manager" ]
