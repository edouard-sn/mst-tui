# Minimal Streamable Torrents CLI

This shit is so not finished

It basically is a daemon that downloads torrents for you and a cli client that does fancy stuff with the daemon.
So u know you can leave the application and it still is downloading wawi mami

This'll probably be in dev for some time cause i'm very lazy

This project is made only for fun, feel free to contribute if you feel like it

## TODO:
- Think about one TorrentClient per torrent.
    - Pros:
        - Different dataDir for each torrent
        - No time to search for torrents
        - No wrapper in main.go
    - Cons:
        - Sounds like an extremely shitty idea
## Goals
- Add/Remove torrents
    - [ ] Daemon
    - [ ] TUI
- Pause/Resume torrent
    - [ ] Daemon
    - [ ] TUI
- Selective files download
    - [ ] Daemon
    - [ ] TUI
- Sequential download per file, maybe per torrent aswell
    - [ ] Daemon
    - [ ] TUI
- (configurable) Choose a software per extension and run the software on chosen file
    - [ ] TUI
- HeaderObfuscationPolicy RequiredPreffered
    - [ ] Daemon

### If not lazy
- Daemon:
    - Remember all torrents + verify data + continue downloading (lol)
    - Option to delete torrent content when deleting torrent
    - InfoHash/Magnet/URL
    - Keybindings
