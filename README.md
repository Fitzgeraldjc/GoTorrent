# GoTorrent

A lightweight BitTorrent client written in Go that supports both traditional torrent files and magnet links.

## Features

- **Torrent File Support**: Download torrents from `.torrent` files
- **Magnet Link Support**: Download torrents from magnet URIs
- **DHT Integration**: Distributed Hash Table support for peer discovery
- **Tracker Support**: HTTP/HTTPS tracker communication
- **Peer-to-Peer Protocol**: Full BitTorrent protocol implementation
- **Multi-peer Downloads**: Concurrent downloading from multiple peers

## Installation

### Prerequisites

- Go 1.24.3 or later

### Build from Source

```bash
git clone https://github.com/Fitzgeraldjc/GoTorrent.git
cd GoTorrent
go build -o gotorrent
```

## Usage

### Basic Command

```bash
./gotorrent <torrent-file-or-magnet-link> <output-path>
```

### Examples

#### Download from Torrent File
```bash
./gotorrent debian-13.0.0-amd64-netinst.iso.torrent ./downloads/debian.iso
```

#### Download from Magnet Link
```bash
./gotorrent "magnet:?xt=urn:btih:HASH&dn=filename&tr=tracker-url" ./downloads/file
```

## Architecture

GoTorrent is organized into several modular packages:

### Core Packages

- **`torrentfile/`** - Torrent file parsing and bencode handling
- **`magnets/`** - Magnet link parsing and processing
- **`p2p/`** - Peer-to-peer protocol implementation
- **`client/`** - BitTorrent client connection management
- **`peers/`** - Peer discovery and management
- **`dht/`** - Distributed Hash Table implementation
- **`handshake/`** - BitTorrent handshake protocol
- **`message/`** - BitTorrent message protocol
- **`bitfield/`** - Piece availability tracking

### Key Components

#### Torrent File Support
- Parses `.torrent` files using bencode encoding
- Extracts tracker URLs, piece hashes, and metadata
- Handles SHA-1 hash verification

#### Magnet Link Support
- Parses magnet URIs with `xt`, `dn`, and `tr` parameters
- Supports hex-encoded info hashes (base32 support planned)
- Integrates with DHT for tracker-less downloads

#### Peer Discovery
- **Tracker-based**: HTTP/HTTPS tracker communication
- **DHT-based**: Distributed peer discovery without trackers
- Automatic fallback between methods

#### Download Engine
- Concurrent piece downloading from multiple peers
- Automatic piece verification using SHA-1 hashes
- Efficient peer connection management
- Request pipelining for optimal throughput

## Protocol Support

- **BitTorrent Protocol**: Full BEP-3 implementation
- **DHT Protocol**: Distributed Hash Table for peer discovery
- **Tracker Protocol**: HTTP/HTTPS tracker communication
- **Magnet Links**: Basic magnet URI support

## Dependencies

- [`github.com/jackpal/bencode-go`](https://github.com/jackpal/bencode-go) - Bencode encoding/decoding

## Technical Details

### Port Configuration
- Default listening port: `6881`
- Configurable in `torrentfile/torrentfile.go:14`

### Performance
- Maximum block size: `16384` bytes
- Maximum request backlog: `50` requests per peer
- Concurrent peer connections for optimal download speed

### Hash Verification
- SHA-1 hash verification for all downloaded pieces
- Automatic re-download of corrupted pieces
- Info hash validation for torrent integrity

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source. Please check the repository for license details.

## Roadmap

- [ ] Base32 encoded info hash support for magnet links
- [ ] Enhanced DHT functionality
- [ ] Resume capability for interrupted downloads
- [ ] Configuration file support
- [ ] Web UI interface
- [ ] Multi-file torrent support improvements

## Troubleshooting

### Common Issues

**No peers found**: Ensure your firewall allows outbound connections on the configured port and that DHT/tracker URLs are accessible.

**Download fails**: Verify the torrent file or magnet link is valid and that the content is still being seeded.

**Permission errors**: Ensure the output directory is writable and you have sufficient disk space.

## Examples in the Wild

The repository includes a sample torrent file (`debian-13.0.0-amd64-netinst.iso.torrent`) for testing purposes.