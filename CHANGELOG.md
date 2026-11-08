# Changelog

All notable changes to Replacearr are documented in this file.

## [0.3.x] - 2026-xx-xx - main (unreleased)

### Added
- Easier mobile navigation withouth auto navigate. Expand first.

### Changed
- Moved discord announcement after docker publish.

### Fixed
- Auth sessions resolve real client IPs behind nginx/Cloudflare (`CF-Connecting-IP`, `True-Client-IP`, `X-Real-IP`, `X-Forwarded-For`) for Settings → Sessions display
- Settings Nav item opens General instead of Media by default.


## [0.3.1] - 2026-11-07 - b548bfc

### Added
- Album cover thumbs in the Music library from Cover Art Archive (cached via mediacover)
- Artist portrait thumbs via Wikidata / Wikimedia Commons with Deezer fallback
- MIT license file
- Discord release announcements and automatic Docker Hub publishes on tagged releases

### Changed
- Music album/artist dialogs keep sticky chrome while scrolling and close with Escape
- Artist grid uses smaller portrait thumbs; image fetches retry with exponential backoff on rate limits

### Fixed
- Production Docker image embeds the real product version from `VERSION` (broken `$$APP_VERSION` ldflags substitution)
- Docker Hub workflow verifies the embedded version after push

## [0.3.0] - 2026-11-06 - 9f8366b

### Added
- Download client discovery (LAN port scan / SSDP / IGD) in Setup Wizard and Queue & Clients
- Native Deluge, Transmission, SABnzbd, and NZBGet download clients
- Test Connection for all built-in download clients
- Expanded movie details: ratings, collection, cast/crew, trailers, and alternate titles
- MediaInfo probing with file details, quality, language, and indexer metadata on media files
- Library Import metadata mapping (TVDb / related IDs) for movies, TV, music, and books
- Import list exclusions
- Recycle Bin settings with scheduled cleanup of deleted media files
- Auth sessions with Settings → Sessions (revoke active logins)
- Per-indexer FlareSolverr enable flag
- Cardigann template compatibility layer and field filters for indexer search
- Indexer connectivity probe APIs
- Left-nav filtering by RBAC permissions

### Changed
- Plugin catalog URL, FlareSolverr URL, and download-client discovery options move from env vars into the UI
- Config paths (`db`, `plugins`, `definitions`) always derive from `REPLACEARR_CONFIG_DIR`; JWT secret is file-backed only
- Download monitor and completed-import handling hardened across client types
- Interactive search and Indexers search UX improved
- Nav group ordering and permission-aware visibility

### Fixed
- Plugin host lifecycle and search/statistics edge cases
- qBittorrent client auth/test path and seeders/peers parsing for Cardigann results

## [0.2.0] - 2026-11-04 - 9e7b778

### Added
- Accordion left navigation for TV, Movies, Music, Books, Settings, and System
- Global Calendar with downloaded / missing / queued / unreleased filters
- Per-library Library Import, Statistics, and Blocklist pages
- Unwanted file filters (global default `.exe`) applied on completed-download import
- System Status health checks, Tasks, Backup, Updates, Events, Logs, and Changelog
- Native APIs for calendar, statistics, blocklist, library import, history, and system ops

### Changed
- Settings categories move into the left-nav accordion (no in-page settings submenu)
- Download client options and indexer options live on Queue & Clients and Indexers tabs
- Backup and update configuration moved under System

## [0.1.0] - 2026-01-01 - 2573d08

### Added
- Unified Sonarr / Radarr / Prowlarr replacement in a single Go binary
- Native Music and Books modules
- Embedded React SPA and plugin host
