# cloud-torrent-dler

Auto-download from a cloud torrent provider (Seedr) and save in a folder (exposed to Plex, for example). Roll your own streaming service!

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
- make folders in Seedr to match your local folder structure
- For example:
  - Seedr Folders:
    - `["Movies/Kids", "Movies/Not Kids", "Shows"]`
  - Local folders would have to match:
  - `/media/DataDrive/`
    - `Shows/`
    - `Movies/`
      - `Kids/`
      - `Not Kids/`
- `go build`
- `./cloud-torrent-dler`
  - use `&` at the end to run in the background
  - or run in a `screen`
  - or add to startup

## Notes

- There is no way to programmatically do this without a paid Seedr account :(
- New episodes added to your ShowRSS feed will have the magnet automatically added to Seedr.
- Anything in the folder list from `config.yaml` will be automatically downloaded to a matching local folder under the `DlRoot` path
- This means that, currently, you will still have to periodically check Seedr and move downloaded files to the appropriate folders.
- Additional features planned
  - Automatically move episodes added via the ShowRSS function to the correct subfolder. This will allow full automation for TV episodes
  - Figuring out how do to something similar for movies.


## TO DO

- Retry downloads if there is an API error. Currently, an error will not result in a download and instead will continue on, deleting the source file. This means you will have to manually re-add things to Seedr on occasion. 
- Better error handling in general. I always put this off even though I know better. 
- Refactor the download function - create a download queue that will be parsed through. This would allow and additional feature of auto-adding shows from ShowRSS directly to the correct folder. Currently, new episodes are added to Seedr, but you manually have to move them to the correct folder. Like an animal. 
