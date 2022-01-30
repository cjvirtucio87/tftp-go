# syntax=docker/dockerfile:1.3-labs
FROM cjvirtucio87/go-devc:rocky8

USER 0

ENV BASH_TOOLS_VERSION=0.1.2 \
    BASH_TOOLS_DIR=/usr/local/lib/bash_tools

RUN --mount=type=cache,id=apt,target=/var/cache/apt <<'EOF'
mkdir --parents "${BASH_TOOLS_DIR}"
curl \
  --location \
  "https://github.com/cjvirtucio87/bash-tools/releases/download/bash-tools-0.1.2/cjvirtucio87-bash-tools-${BASH_TOOLS_VERSION}.tar.gz" \
  | tar \
      --extract \
      --gunzip \
      --directory "${BASH_TOOLS_DIR}"

chgrp --recursive "${REMOTE_GID}" "${BASH_TOOLS_DIR}"
chmod g=u "${BASH_TOOLS_DIR}"

for bin in "${BASH_TOOLS_DIR}/bin"/*; do
  >&2 echo "creating symlink for ${bin}"
  ln -s "${bin}" "/usr/local/bin/$(basename "${bin}")"
done
EOF

USER "${REMOTE_USER}"
