# Base image: https://hub.docker.com/_/golang/
FROM golang:1.17
MAINTAINER Vesa-Pekka Palmu <vpalmu@depili.fi>

# Install golint
ENV GOPATH /go
ENV PATH ${GOPATH}/bin:$PATH
RUN go get -u golang.org/x/lint/golint

# Add apt key for LLVM repository
RUN wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add -

# Add LLVM apt repository
RUN echo "deb http://apt.llvm.org/buster/ llvm-toolchain-buster-7 main" | tee -a /etc/apt/sources.list

# Install clang from LLVM repository and sdl2 headers
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    libsdl2-dev \
    libsdl2-gfx-dev \
    libsdl2-ttf-dev \
    libsdl2-image-dev \
    libsdl2-mixer-dev \
    mingw-w64 \
    nodejs \
    npm \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Set Clang as default CC
ENV set_clang /etc/profile.d/set-clang-cc.sh
RUN echo "export CC=clang-7" | tee -a ${set_clang} && chmod a+x ${set_clang}

RUN npm install -g depili/msi-packager#isolate-mainExecutableFile-error-to-runAfter