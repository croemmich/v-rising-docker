# V Rising Docker

A Docker image for running a V Rising Dedicated Server using Wine64.

## Features

* Custom process manager to ensure clean shutdowns and world auto-saves
* Runs as non-root user (41527)
* Installs/updates the game server on launch
* Tails game server logs to console

## Volumes

The following volumes must be mounted and writeable by the container user.

| Path                    | Description                   |
|-------------------------|-------------------------------|
| /mnt/vrising/server     | Server installation directory |
| /mnt/vrising/persistent | Server config files and saves |

The Steam home directory can optionally be mounted as well to cache steamcmd updates:

| Path        | Description          |
|-------------|----------------------|
| /home/steam | Steam home directory |

## Server Configuration

Server configuration is left as "native" as possible:
1. Stop the server
2. Edit `ServerHostSettings.json` or `ServerGameSettings.json` in your persistent Settings directory.
3. Start the server

For more information on server configuration visit: https://github.com/StunlockStudios/vrising-dedicated-server-instructions#configuring-the-server

## Ports

By default, the following ports are used, however can be changed in `ServerHostSettings.json`:

| Container Port | Type |
|----------------|------|
| 9876           | UDP  |
| 9877           | UDP  |


## Example
```terminal
docker run -d --name='vrising-server' \
--net='bridge' \
--restart=unless-stopped \
-e TZ="America/Chicago" \
-v '/path/on/host/server':'/mnt/vrising/server':'rw' \
-v '/path/on/host/persistent':'/mnt/vrising/persistent':'rw' \
-p 9876:9876/udp \
-p 9877:9877/udp \
'croemmich/v-rising-docker'
```
