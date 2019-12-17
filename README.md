# cloud-torrent-dler

Auto-download from a cloud torrent provider (Seedr) and save in a folder (exposed to Plex, for example)

## Installation

### Pre-reqs

- A paid Seedr account
  - No API or FTP access for free accounts :(
- A ShowRSS.info account (Free! They survive on donations. So do I. Hint hint.)
- Go v1.13+ installed

### Install

- git clone
- `cp config.yaml.template config.yaml`
- edit `config.yaml` with your parameters
- make a `Completed` folder in Seedr (to prevent downloading then deleting in-progress downloads)
- `go build -i`
- `./cloud-torrent-dler` us `&` at the end to run in the background, or run in a `screen`

## Notes

- There is no way to programmatically do this without a paid Seedr account :(
- New episodes added to your ShowRSS feed will have the magnet automatically added to Seedr.
- Anything in the `Completed` folder from `config.yaml` will be automatically downloaded
- Additional features planned
  - Multiple completed folders mapped to multiple download destinations
  - Auto deletion from Seedr once download completes (coming soon)
