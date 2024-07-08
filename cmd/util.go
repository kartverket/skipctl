package cmd

import (
	"os"
	"strings"

	"github.com/kartverket/skipctl/pkg/discovery"
	"github.com/spf13/cobra"
)

var activeApiServer discovery.ApiServer

func ValidateAPIServerName(_ *cobra.Command, _ []string) {
	if len(apiServer) == 0 {
		log.Error("no api server specified, exiting")
		os.Exit(1)
	}

	var matchFound = false
	for _, server := range apiServers {
		if strings.ToLower(apiServer) == strings.ToLower(server.Name) {
			matchFound = true
			activeApiServer = server
			break
		}
	}

	if !matchFound {
		var names []string
		for _, server := range apiServers {
			names = append(names, strings.ToLower(server.Name))
		}

		log.Error("unknown api server - please pick another supported", "specified", apiServer, "supported", names)
		os.Exit(1)
	}
}
