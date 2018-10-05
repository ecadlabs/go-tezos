package tezos

import (
	"context"
	"fmt"
	"net/http"
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
	path := fmt.Sprintf("%s/%s", s.client.BaseURL, "/network/stat")

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
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

//NetworkStats contains Global network bandwidth totals and usage in B/s
type NetworkStats struct {
	TotalBytesSent int64 `json:"total_sent,string"`
	TotalBytesRecv int64 `json:"total_recv,string"`
	CurrentInflow  int64 `json:"current_inflow"`
	CurrentOutflow int64 `json:"current_outflow"`
}
