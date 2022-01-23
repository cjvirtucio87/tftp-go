FROM rockylinux/rockylinux

USER 0

ENV APP_DIR=/opt/app \
    CLCONF_VERSION=2.1.5

RUN set -e; \
    dnf -y update; \
    dnf -y group install development; \
    mkdir --parents "${APP_DIR}"; \
    chgrp 0 "${APP_DIR}"; \
    chmod g=u "${APP_DIR}"; \
    : ;

USER 1001
