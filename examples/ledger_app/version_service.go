package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/go-daml/pkg/client"
	"github.com/smartcontractkit/go-daml/pkg/model"
)

func RunVersionService(cl *client.DamlBindingClient) {
	req := &model.GetLedgerAPIVersionRequest{}

	version, err := cl.VersionService.GetLedgerAPIVersion(context.Background(), req)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get ledger API version")
	}

	log.Info().
		Str("version", version.Version).
		Interface("features", version.Features).
		Msg("ledger API version")
}
