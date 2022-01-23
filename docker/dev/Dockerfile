# syntax=docker/dockerfile:1.3-labs
FROM cjvirtucio87/docker-base-rockylinux:latest

USER 0

ENV USR_LOCAL_GO=/usr/local/go
ENV PATH="${USR_LOCAL_GO}/bin:${PATH}"

RUN <<'EOF'
set -e

mkdir --parents "${USR_LOCAL_GO}"
curl \
  --location \
  https://go.dev/dl/go1.17.6.linux-amd64.tar.gz \
  | tar \
      --extract \
      --gunzip \
      --strip-components 1 \
      --directory "${USR_LOCAL_GO}"

chgrp --recursive 0 "${USR_LOCAL_GO}"
chmod --recursive 775 "${USR_LOCAL_GO}"

go version
EOF

COPY docker/dev/contrib /

USER 1001

ENTRYPOINT ["/docker-entrypoint.sh"]
