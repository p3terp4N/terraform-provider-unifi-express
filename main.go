package main // import "github.com/p3terp4N/terraform-provider-unifi-express"

import (
	"context"
	"flag"
	"log"

	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// Remove any date and time prefix in log package function output to
	// prevent duplicate timestamp and incorrect log level setting
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/p3terp4N/unifi-express",
		Debug:   debugMode,
	}

	err := providerserver.Serve(context.Background(), provider.NewV2(version), opts)
	if err != nil {
		log.Fatal(err)
	}
}
