# cloud-torrent-dler

Auto-download from a cloud torrent provider (Seedr) and save in a folder (exposed to Plex, for example)

## Installation

- git clone
- `cp config.yaml.template config.yaml`
- edit `config.yaml` with your parameters
- `go run main.go` - or throw it in a cron and continually download new items

Note: you will need a `Master` Seedr account for API access. There is FTP access for lower plans, and an FTP version of this is _"on the roadmap"_.
