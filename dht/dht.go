package dht

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/Fitzgeraldjc/GoTorrent/peers"
)

type DHT struct {
	nodeID [20]byte
	conn   *net.UDPConn
}

func New(nodeID [20]byte) (*DHT, error) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &DHT{
		nodeID: nodeID,
		conn:   conn,
	}, nil
}

func (d *DHT) FindPeers(infoHash [20]byte) ([]peers.Peer, error) {
	bootstrapNodes := []string{
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881",
	}

	var allPeers []peers.Peer
	
	for _, node := range bootstrapNodes {
		peerList, err := d.queryNode(node, infoHash)
		if err != nil {
			continue
		}
		allPeers = append(allPeers, peerList...)
	}

	return allPeers, nil
}

func (d *DHT) queryNode(nodeAddr string, infoHash [20]byte) ([]peers.Peer, error) {
	addr, err := net.ResolveUDPAddr("udp", nodeAddr)
	if err != nil {
		return nil, err
	}

	query := d.buildGetPeersQuery(infoHash)
	
	d.conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = d.conn.WriteToUDP(query, addr)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	n, _, err := d.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}

	return d.parseGetPeersResponse(buf[:n])
}

func (d *DHT) buildGetPeersQuery(infoHash [20]byte) []byte {
	txnID := []byte("aa")
	
	query := []byte("d1:ad2:id20:")
	query = append(query, d.nodeID[:]...)
	query = append(query, []byte("9:info_hash20:")...)
	query = append(query, infoHash[:]...)
	query = append(query, []byte("e1:q9:get_peers1:t2:")...)
	query = append(query, txnID...)
	query = append(query, []byte("1:y1:qe")...)
	
	return query
}

func (d *DHT) parseGetPeersResponse(data []byte) ([]peers.Peer, error) {
	var peerList []peers.Peer
	
	valuesStart := findBytes(data, []byte("6:valuesl"))
	if valuesStart == -1 {
		return nil, fmt.Errorf("no peers found in DHT response")
	}
	
	offset := valuesStart + len("6:valuesl")
	
	for offset < len(data) && data[offset] != 'e' {
		if data[offset] == '6' && offset+1 < len(data) && data[offset+1] == ':' {
			offset += 2
			if offset+6 > len(data) {
				break
			}
			
			ip := net.IP(data[offset:offset+4])
			port := binary.BigEndian.Uint16(data[offset+4:offset+6])
			
			peer := peers.Peer{
				IP:   ip,
				Port: port,
			}
			peerList = append(peerList, peer)
			offset += 6
		} else {
			offset++
		}
	}
	
	return peerList, nil
}

func findBytes(haystack, needle []byte) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func (d *DHT) Close() error {
	return d.conn.Close()
}