# cloud-torrent-dler

Auto-download from a cloud torrent provider (Seedr) and save in a folder (exposed to Plex, for example)

## Installation

### Pre-reqs

- A paid [Seedr.cc](https://www.seedr.cc/?r=211) account
  - No API or FTP access for free accounts :(
- A ShowRSS.info account (Free! They survive on donations. So do I. [Hint hint.](paypal.me/jdale215))
- Go v1.13+ installed

### Install

- git clone
- `cp config.yaml.template config.yaml`
- edit `config.yaml` with your parameters
- make folder in Seedr to match your local folder structure
- For example:
  - Seedr Folders:
    - `["Movies/Kids", "Movies/Not Kids", "Shows"]`
  - Local folders would have to match:
  - `/media/DataDrive/`
    - `Shows/`
    - `Movies/`
      - `Kids/`
      - `Not Kids/`
- `go build -i`
- `./cloud-torrent-dler`
  - use `&` at the end to run in the background
  - or run in a `screen`
  - or add to startup

## Notes

- There is no way to programmatically do this without a paid Seedr account :(
- New episodes added to your ShowRSS feed will have the magnet automatically added to Seedr.
- Anything in the folder list from `config.yaml` will be automatically downloaded to a matching local folder under the `DlRoot` path
- Additional features planned
  - Auto deletion from Seedr once download completes (coming soon)
