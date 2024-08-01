package discovery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/kartverket/skipctl/pkg/constants"
)

var (
	resolver = net.DefaultResolver
	b64      = base64.StdEncoding
)

type APIServer struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

// DiscoverAPIServers will try to do a TXT lookup for a given DNS name. If found it will
// attempt a unmarshal(base64_decode(TXT_RECORD_VALUE)) (pseudo code) into a list of APIServer
// structs.
//
// One TXT record per server.
func DiscoverAPIServers(dnsKey string) ([]APIServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DNSDiscoverTimeout)
	defer cancel()

	records, err := resolver.LookupTXT(ctx, dnsKey)
	if err != nil {
		return nil, fmt.Errorf("failed discover available API servers: %w", err)
	}

	var apiServers []APIServer

	for _, txtRecord := range records {
		// Wrap the outer layer
		decodedBytes, decodeErr := b64.DecodeString(txtRecord)
		if decodeErr != nil {
			return nil, fmt.Errorf("failed base64 decoding TXT record: %w", decodeErr)
		}

		// Decode JSON into a usable structure
		var apiServer APIServer
		err = json.Unmarshal(decodedBytes, &apiServer)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshalling TXT record: %w", err)
		}

		apiServers = append(apiServers, apiServer)
	}

	return apiServers, nil
}
