# Multi-stage build: compile static binaries + download pre-built tools
# Stages run in parallel via BuildKit for faster builds

FROM alpine:3.21 AS base

RUN apk add --no-cache \
    build-base cmake git curl wget linux-headers \
    ncurses-dev ncurses-static \
    go ninja gettext-tiny-dev unzip autoconf automake libtool pkgconf \
    zlib-dev zlib-static \
    openssl-dev openssl-libs-static \
    curl-dev curl-static \
    expat-dev expat-static \
    pcre2-dev perl \
    nghttp2-static brotli-static \
    libidn2-static libunistring-static libpsl-static zstd-static

WORKDIR /build

# ============================================================
# git — static with HTTPS support
# ============================================================
FROM base AS git-build
ARG GIT_VERSION=2.53.0
RUN curl -fsSL "https://github.com/git/git/archive/refs/tags/v${GIT_VERSION}.tar.gz" | tar xz && \
    cd git-${GIT_VERSION} && \
    echo "prefix = /opt/git" > config.mak && \
    echo "NO_TCLTK = YesPlease" >> config.mak && \
    echo "NO_GETTEXT = YesPlease" >> config.mak && \
    echo "NO_PERL = YesPlease" >> config.mak && \
    echo "NO_PYTHON = YesPlease" >> config.mak && \
    echo "NO_EXPAT = YesPlease" >> config.mak && \
    echo "NO_NSEC = YesPlease" >> config.mak && \
    echo "NO_REGEX = YesPlease" >> config.mak && \
    echo "CURL_LDFLAGS = $(pkg-config --static --libs libcurl | sed 's/-ldl//g')" >> config.mak && \
    echo "CFLAGS = -Os -DNDEBUG" >> config.mak && \
    echo "LDFLAGS = -static -Wl,--allow-multiple-definition" >> config.mak && \
    make -j$(nproc) && \
    make install && \
    strip /opt/git/bin/git /opt/git/bin/git-remote-http*

# ============================================================
# zsh — static with essential modules
# ============================================================
FROM base AS zsh-build
ARG ZSH_VERSION=5.9
RUN git clone --depth 1 --branch zsh-${ZSH_VERSION} https://github.com/zsh-users/zsh.git && \
    cd zsh && \
    ./Util/preconfig && \
    ./configure \
      --prefix=/opt/zsh \
      --enable-static \
      --disable-dynamic \
      --enable-multibyte \
      --with-tcsetpgrp \
      LDFLAGS="-static" \
      CFLAGS="-Os -DNDEBUG" && \
    for mod in compctl complete complist computil zle zutil parameter terminfo datetime stat system mathfunc "net/socket" "net/tcp"; do \
      escaped=$(echo "$mod" | sed 's|/|\\/|g'); \
      sed -i "/name=zsh\/${escaped}/s/link=no/link=static/; /name=zsh\/${escaped}/s/load=no/load=yes/" config.modules; \
    done && \
    make -j$(nproc) && \
    make install.bin install.fns && \
    strip /opt/zsh/bin/zsh

# ============================================================
# htop — static
# ============================================================
FROM base AS htop-build
ARG HTOP_VERSION=3.4.1
RUN git clone --depth 1 --branch ${HTOP_VERSION} https://github.com/htop-dev/htop.git && \
    cd htop && \
    ./autogen.sh && \
    ./configure --enable-static LDFLAGS="-static" CFLAGS="-Os -DNDEBUG" && \
    make -j$(nproc) && \
    strip htop

# ============================================================
# neovim — static
# ============================================================
FROM base AS nvim-build
RUN git clone --depth 1 --branch stable https://github.com/neovim/neovim.git && \
    cd neovim && \
    make CMAKE_BUILD_TYPE=Release \
         CMAKE_EXTRA_FLAGS="-DCMAKE_INSTALL_PREFIX=/opt/nvim -DCMAKE_EXE_LINKER_FLAGS=-static" \
         -j$(nproc) && \
    make install && \
    strip /opt/nvim/bin/nvim

# ============================================================
# Download pre-built binaries + assemble tarball
# ============================================================
FROM ubuntu:24.04 AS assembler

RUN apt-get update && apt-get install -y --no-install-recommends \
      curl ca-certificates unzip file \
    && rm -rf /var/lib/apt/lists/*

COPY versions.env /tmp/versions.env
COPY scripts/download-binaries.sh /tmp/scripts/download-binaries.sh

# Download pre-built binaries (skip nvim — we built it from source)
RUN SKIP_NVIM=1 bash /tmp/scripts/download-binaries.sh /staging linux

# Add compiled static binaries
COPY --from=git-build /opt/git /staging/git/
COPY --from=zsh-build /opt/zsh /staging/zsh/
COPY --from=htop-build /build/htop/htop /staging/bin/htop
COPY --from=nvim-build /opt/nvim /staging/nvim/

# Create symlinks in bin/
RUN ln -sf ../git/bin/git /staging/bin/git && \
    ln -sf ../zsh/bin/zsh /staging/bin/zsh && \
    ln -sf ../nvim/bin/nvim /staging/bin/nvim && \
    ln -sf ../go/bin/go /staging/bin/go && \
    ln -sf ../go/bin/gofmt /staging/bin/gofmt && \
    chmod +x /staging/bin/*

# Verify all binaries are static
RUN echo "==> Verifying binaries:" && \
    for f in /staging/bin/git /staging/bin/zsh /staging/bin/htop /staging/nvim/bin/nvim; do \
      echo "  $(basename $f): $(file $f | grep -o 'statically linked' || echo 'dynamically linked')"; \
    done

# Generate checksums
RUN find /staging -type f -executable | sort | xargs sha256sum > /staging/SHA256SUMS

# Package
RUN tar czf /dotpack.tar.gz -C /staging .

CMD ["cat", "/dotpack.tar.gz"]
