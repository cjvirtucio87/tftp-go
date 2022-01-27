# syntax=docker/dockerfile:1.3-labs
FROM cjvirtucio87/docker-base-rockylinux:latest

USER 0

ENV USR_LOCAL_GO=/usr/local/go
ENV PATH="${USR_LOCAL_GO}/bin:${PATH}"
ENV USER=dev
ENV HOME="/home/${USER}"

RUN <<'EOF'
set -e

mkdir --parents "${HOME}"
chmod g=u /etc/passwd "${HOME}"

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
chmod --recursive g=u "${USR_LOCAL_GO}"

go version
EOF

COPY docker/dev/contrib /

RUN <<'EOF'
chgrp 0 /docker-entrypoint.sh
chmod g=u /docker-entrypoint.sh
EOF

USER "${USER}"

WORKDIR "${HOME}"

ENTRYPOINT ["/docker-entrypoint.sh"]
