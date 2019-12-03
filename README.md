# cloud-torrent-dler

Auto-download from a cloud torrent provider (Seedr) and save in a folder (exposed to Plex, for example)

## Installation

- git clone
- `cp config.yaml.template config.yaml`
- edit `config.yaml` with your parameters
- make a `Completed` folder in Seedr (to prevent downloading then deleting in-progress downloads)
- `go run *.go`
- There is no way to programmatically do this without a paid Seedr account :(
