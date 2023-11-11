FROM golang:1.21.4-bullseye as vrpm
COPY vrpm /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -v -o vrpm && chmod +x vrpm

FROM ubuntu:22.04
LABEL maintainer="Chris Roemmich"
VOLUME ["/mnt/vrising/server", "/mnt/vrising/persistent"]

ARG DEBIAN_FRONTEND="noninteractive"
ARG STEAM_USER_ID=41527
ARG STEAM_GROUP_ID=41527

# Prepare
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y ca-certificates \
       curl \
       locales \
       tzdata \
       wget && \
    rm -rf /var/lib/apt/lists/* && \
    locale-gen --no-purge en_US.UTF-8

ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

# Install steamcmd
RUN dpkg --add-architecture i386 && \
    apt-get update && \
    echo steam steam/question select "I AGREE" | debconf-set-selections && \
    echo steam steam/license note '' | debconf-set-selections && \
    apt-get install -y steamcmd && \
    rm -rf /var/lib/apt/lists/* && \
    ln -s /usr/games/steamcmd /usr/bin/steamcmd

# Create a Steam user
RUN addgroup --system --gid ${STEAM_GROUP_ID} steam && \
    adduser --system --uid ${STEAM_USER_ID} --gid ${STEAM_GROUP_ID} \
        --home /home/steam \
        --shell /usr/sbin/nologin \
        --disabled-password steam
ENV HOME /home/steam

## Install Wine and X-Server
RUN apt-get update && \
    apt-get install -y wine \
        xserver-xorg \
        xvfb && \
    rm -rf /var/lib/apt/lists/*

COPY --from=vrpm /build/vrpm /usr/local/bin/vrpm

WORKDIR /mnt/vrising/server

USER ${STEAM_USER_ID}

EXPOSE 9876/udp
EXPOSE 9877/udp

CMD ["/usr/local/bin/vrpm"]
