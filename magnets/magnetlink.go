package magnets

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Fitzgeraldjc/GoTorrent/dht"
	"github.com/Fitzgeraldjc/GoTorrent/p2p"
	"github.com/Fitzgeraldjc/GoTorrent/peers"
	"github.com/Fitzgeraldjc/GoTorrent/torrentfile"
)

type MagnetLink struct {
	InfoHash  [20]byte
	Name      string
	Trackers  []string
	Length    int
}

func Parse(magnetURI string) (*MagnetLink, error) {
	if !strings.HasPrefix(magnetURI, "magnet:?") {
		return nil, fmt.Errorf("invalid magnet URI: must start with magnet:?")
	}

	u, err := url.Parse(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse magnet URI: %v", err)
	}

	params := u.Query()
	
	xtParams := params["xt"]
	if len(xtParams) == 0 {
		return nil, fmt.Errorf("magnet URI missing exact topic (xt) parameter")
	}

	var infoHashStr string
	for _, xt := range xtParams {
		if strings.HasPrefix(xt, "urn:btih:") {
			infoHashStr = strings.TrimPrefix(xt, "urn:btih:")
			break
		}
	}

	if infoHashStr == "" {
		return nil, fmt.Errorf("magnet URI missing BitTorrent info hash")
	}

	var infoHash [20]byte
	if len(infoHashStr) == 40 {
		hashBytes, err := hex.DecodeString(infoHashStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hex info hash: %v", err)
		}
		copy(infoHash[:], hashBytes)
	} else if len(infoHashStr) == 32 {
		return nil, fmt.Errorf("base32 encoded info hashes not yet supported")
	} else {
		return nil, fmt.Errorf("invalid info hash length: expected 40 hex chars or 32 base32 chars")
	}

	magnet := &MagnetLink{
		InfoHash: infoHash,
		Trackers: params["tr"],
	}

	if len(params["dn"]) > 0 {
		magnet.Name = params["dn"][0]
	}

	return magnet, nil
}

func (m *MagnetLink) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	allPeers, err := m.findPeers(peerID)
	if err != nil {
		return fmt.Errorf("failed to find peers: %v", err)
	}

	if len(allPeers) == 0 {
		return fmt.Errorf("no peers found for magnet link")
	}

	torrent := p2p.Torrent{
		Peers:       allPeers,
		PeerID:      peerID,
		InfoHash:    m.InfoHash,
		PieceHashes: nil, // Will be retrieved from peers
		PieceLength: 0,   // Will be retrieved from peers
		Length:      0,   // Will be retrieved from peers  
		Name:        m.Name,
	}

	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (m *MagnetLink) findPeers(peerID [20]byte) ([]peers.Peer, error) {
	var allPeers []peers.Peer

	if len(m.Trackers) > 0 {
		trackerPeers, err := m.requestTracker(peerID)
		if err == nil {
			allPeers = append(allPeers, trackerPeers...)
		}
	}

	dhtNode, err := dht.New(peerID)
	if err == nil {
		defer dhtNode.Close()
		dhtPeers, err := dhtNode.FindPeers(m.InfoHash)
		if err == nil {
			allPeers = append(allPeers, dhtPeers...)
		}
	}

	return allPeers, nil
}

func (m *MagnetLink) requestTracker(peerID [20]byte) ([]peers.Peer, error) {
	if len(m.Trackers) == 0 {
		return nil, fmt.Errorf("no trackers available")
	}

	tf := &torrentfile.TorrentFile{
		Announce: m.Trackers[0],
		InfoHash: m.InfoHash,
		Length:   m.Length,
	}

	return tf.RequestPeers(peerID, torrentfile.Port)
}
