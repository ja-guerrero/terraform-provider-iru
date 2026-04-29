package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/ja-guerrero/terraform-provider-iru/internal/provider"
)

// Run the docs generation tool, check its documentation for more information on how it works:
// http://github.com/hashicorp/terraform-plugin-docs
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

// These variables are set by goreleaser at build time.
var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "github.com/ja-guerrero/iru",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
