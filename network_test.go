package tezos

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNetworkStats(t *testing.T) {
	expected := &NetworkStats{
		TotalBytesSent: 291690080,
		TotalBytesRecv: 532639553,
		CurrentInflow:  23596,
		CurrentOutflow: 14972,
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/network/stat", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)

		buf, err := ioutil.ReadFile("fixtures/network/stat.json")
		require.NoError(t, err, "error reading fixture")

		_, err = w.Write(buf)
		require.NoError(t, err, "error writing HTTP response")
	}))
	defer s.Close()

	c, err := NewClient(nil, s.URL)
	require.NoError(t, err, "error creating client")

	stats, err := c.Network.GetStats(context.Background())
	require.NoError(t, err, "error getting network stats")

	require.Equal(t, expected, stats, "unexpected stats response")
}

func TestGetNetworkConnections(t *testing.T) {
	expected := []NetworkConnection{NetworkConnection{Incoming: false, PeerID: "idt5qvkLiJ15rb6yJU1bjpGmdyYnPJ", IDPoint: NetworkIDPoint{Addr: "::ffff:34.253.64.43", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []NetworkVersion{NetworkVersion{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: false, LocalMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}, RemoteMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}}, NetworkConnection{Incoming: true, PeerID: "ids8VJTHEuyND6B8ahGgXPAJ3BDp1c", IDPoint: NetworkIDPoint{Addr: "::ffff:176.31.255.202", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []NetworkVersion{NetworkVersion{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: true, LocalMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}, RemoteMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}}}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/network/connections", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)

		buf, err := ioutil.ReadFile("fixtures/network/connections.json")
		require.NoError(t, err, "error reading fixture")

		_, err = w.Write(buf)
		require.NoError(t, err, "error writing HTTP response")
	}))
	defer s.Close()

	c, err := NewClient(nil, s.URL)
	require.NoError(t, err, "error creating client")

	conns, err := c.Network.GetConnections(context.Background())
	require.NoError(t, err, "error getting network connections")

	require.Equal(t, expected, conns, "unexpected network connections response")
}
