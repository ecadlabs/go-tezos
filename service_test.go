package tezos

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceGetMethods(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		get           func(s *Service) (interface{}, error)
		respFixture   string
		respStatus    int
		expectedPath  string
		expectedValue interface{}
		errMsg        string
	}{
		{
			get:          func(s *Service) (interface{}, error) { return s.GetStats(ctx) },
			respFixture:  "fixtures/network/stat.json",
			expectedPath: "/network/stat",
			expectedValue: &NetworkStats{
				TotalBytesSent: 291690080,
				TotalBytesRecv: 532639553,
				CurrentInflow:  23596,
				CurrentOutflow: 14972,
			},
		},
		{
			get:           func(s *Service) (interface{}, error) { return s.GetConnections(ctx) },
			respFixture:   "fixtures/network/connections.json",
			expectedPath:  "/network/connections",
			expectedValue: []NetworkConnection{NetworkConnection{Incoming: false, PeerID: "idt5qvkLiJ15rb6yJU1bjpGmdyYnPJ", IDPoint: NetworkIDPoint{Addr: "::ffff:34.253.64.43", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []NetworkVersion{NetworkVersion{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: false, LocalMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}, RemoteMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}}, NetworkConnection{Incoming: true, PeerID: "ids8VJTHEuyND6B8ahGgXPAJ3BDp1c", IDPoint: NetworkIDPoint{Addr: "::ffff:176.31.255.202", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []NetworkVersion{NetworkVersion{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: true, LocalMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}, RemoteMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetDelegateBalance(ctx, "main", "head", "tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5")
			},
			respFixture:   "fixtures/contract/delegate_balance.json",
			expectedPath:  "/chains/main/blocks/head/context/delegates/tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5/balance",
			expectedValue: "13490453135591",
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetContractBalance(ctx, "main", "head", "tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5")
			},
			respFixture:   "fixtures/contract/contract_balance.json",
			expectedPath:  "/chains/main/blocks/head/context/contracts/tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5/balance",
			expectedValue: "4700354460878",
		},
		// Handling 5xx errors from the Tezos node with RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:   500,
			respFixture:  "fixtures/error.json",
			expectedPath: "/network/stat",
			errMsg:       `Tezos RPC error (kind = "permanent", id = "proto.002-PsYLVpVv.context.storage_error")`,
		},
		// Handling 5xx errors from the Tezos node with empty RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:   500,
			respFixture:  "fixtures/empty_error.json",
			expectedPath: "/network/stat",
			errMsg:       `received a Tezos RPC error response with 0 errors`,
		},
		// Handling 5xx errors from the Tezos node with malformed RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:   500,
			respFixture:  "fixtures/malformed_error.json",
			expectedPath: "/network/stat",
			errMsg:       `error decoding Tezos RPC errors: invalid character ',' looking for beginning of value`,
		},
		// Handling unexpected response status codes.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetStats(ctx)
			},
			respStatus:   404,
			respFixture:  "fixtures/empty.json",
			expectedPath: "/network/stat",
			errMsg:       `unexpected response status: 404 Not Found`,
		},
	}

	for _, test := range tests {
		// Start a test HTTP server that responds as specified in the test case parameters.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, test.expectedPath, r.URL.Path)
			require.Equal(t, http.MethodGet, r.Method)

			buf, err := ioutil.ReadFile(test.respFixture)
			require.NoError(t, err, "error reading fixture %q", test.respFixture)

			if test.respStatus != 0 {
				w.WriteHeader(test.respStatus)
			}
			_, err = w.Write(buf)
			require.NoError(t, err, "error writing HTTP response")
		}))

		c, err := NewRPCClient(nil, srv.URL)
		require.NoError(t, err, "error creating client")

		s := &Service{Client: c}

		value, err := test.get(s)
		if test.errMsg == "" {
			require.NoError(t, err, "error getting value")
			require.Equal(t, test.expectedValue, value, "unexpected value")
		} else {
			require.EqualError(t, err, test.errMsg, "unexpected error string")
		}

		srv.Close()
	}
}
