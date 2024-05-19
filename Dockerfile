FROM golang:1.22.3-bookworm as vrpm
COPY vrpm /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -v -o vrpm && chmod +x vrpm

FROM ubuntu:24.04
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
RUN groupadd --gid ${STEAM_GROUP_ID} steam && \
    useradd --uid ${STEAM_USER_ID} --gid ${STEAM_GROUP_ID} \
      --home-dir /home/steam \
      --create-home \
      --shell /usr/sbin/nologin \
      steam
ENV HOME /home/steam

## Install Wine and X-Server
RUN mkdir -pm755 /etc/apt/keyrings && \
    wget -O /etc/apt/keyrings/winehq-archive.key https://dl.winehq.org/wine-builds/winehq.key && \
    wget -NP /etc/apt/sources.list.d/ https://dl.winehq.org/wine-builds/ubuntu/dists/noble/winehq-noble.sources && \
    apt-get update && \
    apt-get install -y winehq-devel \
        xserver-xorg \
        xvfb && \
    rm -rf /var/lib/apt/lists/*

COPY --from=vrpm /build/vrpm /usr/local/bin/vrpm

WORKDIR /mnt/vrising/server

USER ${STEAM_USER_ID}

EXPOSE 9876/udp
EXPOSE 9877/udp

CMD ["/usr/local/bin/vrpm"]
