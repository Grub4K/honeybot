# Honeypot bot

Go Discord Bot for setting up a honeypot channel

## Usage
Create a `config.json` in the programs working directory with the following format.
```json
{
    "token": "<bot token>",
    "channels": {
        "<channel id>": {
            "delete": "24h",
            "ignored_roles": [
                "<role id>",
                "<role id>",
            ]
        }
    }
}
```

Currently the `delete` key only supports increments of 24h.
Any changes to the config file will be automatically watched for using a filesystem watcher.

### Docker Compose
Use a compose file that uses the published image from the GitHub Container Registry:
```yaml
services:
    honeybot:
        image: ghcr.io/grub4k/honeybot
        volumes:
            - ./config.json:/config.json:ro
```
