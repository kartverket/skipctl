package discovery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

var resolver = net.DefaultResolver

type ApiServer struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

func DiscoverAPIServers(dnsKey string) ([]ApiServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	records, err := resolver.LookupTXT(ctx, dnsKey)
	if err != nil {
		return nil, fmt.Errorf("failed discover available API servers: %w", err)
	}

	if len(records) > 1 {
		return nil, fmt.Errorf("found more than one TXT record with the same name: %s", dnsKey)
	}

	// Wrap the outer layer
	decodedBytes, err := base64.StdEncoding.DecodeString(records[0])
	if err != nil {
		return nil, fmt.Errorf("failed base64 decoding TXT record: %w", err)
	}

	// Decode JSON into a usable structure
	var apiServers []ApiServer
	err = json.Unmarshal(decodedBytes, &apiServers)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling TXT record: %w", err)
	}

	return apiServers, nil
}
