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

var _ NetworkService = &NetworkServiceOp{}

// NetworkStats contains Global network bandwidth totals and usage in B/s.
type NetworkStats struct {
	TotalBytesSent int64 `json:"total_sent,string"`
	TotalBytesRecv int64 `json:"total_recv,string"`
	CurrentInflow  int64 `json:"current_inflow"`
	CurrentOutflow int64 `json:"current_outflow"`
}
