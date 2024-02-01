# Minimal Streamable Torrents TUI

This shit is so not finished

It basically is a daemon that downloads torrents for you and a cli client that does fancy stuff with the daemon.
So u know you can leave the application and it still is downloading wawi mami

This'll probably be in dev for some time cause i'm very lazy

This project is made only for fun, feel free to contribute if you feel like it

## Goals
- Add/Remove torrents
    - [X] Daemon
    - [ ] TUI
- Pause/Resume torrent - use AllowDataDownload and store the state in the repo
    - [ ] Daemon
    - [ ] TUI
- Selective files download
    - [X] Daemon
    - [ ] TUI
- Sequential download per file, maybe per torrent aswell, and find a way to cancel it
    - [X] Daemon - Tested and works for files, need canceling now
    - [ ] TUI
- (configurable) Choose a software per extension and run the software on chosen file
    - [ ] TUI

### If not lazy
- Daemon:
    - Remember all torrents + verify data + continue downloading (lol)
    - Option to delete torrent content when deleting torrent
    - InfoHash/Magnet/URL
    - Keybindings
