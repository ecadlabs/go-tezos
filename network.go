package tezos

import (
	"context"
	"net/http"
	"path"
)

// NetworkService is an interface for retrieving network stats from the tezos RPC api
// https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat
type NetworkService interface {
	GetStats(context.Context) (*NetworkStats, error)
	GetConnections(context.Context) ([]NetworkConnection, error)
}

// NetworkServiceOp handles communication with the `/network` related tezos RPC methods
type NetworkServiceOp struct {
	client *Client
}

// GetStats returns current network stats https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat
func (s *NetworkServiceOp) GetStats(ctx context.Context) (*NetworkStats, error) {
	url := *s.client.BaseURL
	url.Path = path.Join(url.Path, "/network/stat")

	req, err := s.client.NewRequest(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	stats := new(NetworkStats)
	//TODO make use of response, only works for GETs
	_, err = s.client.Do(ctx, req, stats)
	if err != nil {
		return nil, err
	}
	return stats, err
}

// GetConnections returns all network connections http://tezos.gitlab.io/mainnet/api/rpc.html#get-network-connections
func (s *NetworkServiceOp) GetConnections(ctx context.Context) ([]NetworkConnection, error) {
	url := *s.client.BaseURL
	url.Path = path.Join(url.Path, "/network/connections")

	req, err := s.client.NewRequest(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	conns := []NetworkConnection{}
	_, err = s.client.Do(ctx, req, &conns)
	if err != nil {
		return nil, err
	}
	return conns, err
}

var _ NetworkService = &NetworkServiceOp{}

// NetworkStats contains Global network bandwidth totals and usage in B/s.
type NetworkStats struct {
	TotalBytesSent int64 `json:"total_sent,string"`
	TotalBytesRecv int64 `json:"total_recv,string"`
	CurrentInflow  int64 `json:"current_inflow"`
	CurrentOutflow int64 `json:"current_outflow"`
}

type NetworkConnection struct {
	Incoming         bool             `json:"incoming"`
	PeerID           string           `json:"peer_id"`
	IDPoint          NetworkIDPoint   `json:"id_point"`
	RemoteSocketPort uint16           `json:"remote_socket_port"`
	Versions         []NetworkVersion `json:"versions"`
	Private          bool             `json:"private"`
	LocalMetadata    NetworkMetadata  `json:"local_metadata"`
	RemoteMetadata   NetworkMetadata  `json:"remote_metadata"`
}

type NetworkIDPoint struct {
	Addr string `json:"addr"`
	Port uint16 `json:"port"`
}

type NetworkVersion struct {
	Name  string `json:"name"`
	Major uint16 `json:"major"`
	Minor uint16 `json:"minor"`
}

type NetworkMetadata struct {
	DisableMempool bool `json:"disable_mempool"`
	PrivateNode    bool `json:"private_node"`
}
