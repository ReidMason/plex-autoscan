# Plex autoscan

A simple server that listens to Sonarr webhooks and triggers a Plex rescan on specific directories that have been updated.

For a more fully featured implementation use [autoscan](https://github.com/Cloudbox/autoscan). For me this was slower than I wanted so I wrote this version to remove features and keep the process simple.

## Deployment

Example docker compose file

```yml
version: "3.8"

services:
  plex-autoscan:
    container_name: plex-autoscan
    image: ghcr.io/reidmason/plex-autoscan:latest
    restart: always
    ports:
      - 3030:3030
    volumes:
      - "./appdata/data:/data"
```

Create a config file `config.json` inside the data directory\
Example config

```json
{
  "PlexHost": "http://192.168.1.1",
  "PlexToken": "PLEX-TOKEN",
  "PlexPort": 32400,
  "Remappings": {
    "sonarr": [
      {
        "from": "/tv",
        "to": "/data"
      }
    ]
  }
}
```
