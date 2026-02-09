package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/go-daml/pkg/client"
)

func getAvailableUser(cl *client.DamlBindingClient) string {
	users, err := cl.UserMng.ListUsers(context.Background())
	if err != nil || len(users) == 0 {
		log.Warn().Err(err).Msg("failed to list users, using default")
		return "participant_admin"
	}
	return users[0].ID
}

func getAvailableParty(cl *client.DamlBindingClient) string {
	response, err := cl.PartyMng.ListKnownParties(context.Background(), "", 10, "")
	if err != nil || len(response.PartyDetails) == 0 {
		log.Warn().Err(err).Msg("failed to list parties, using default")
		return "participant_admin"
	}
	return response.PartyDetails[0].Party
}

func getAvailableUserAndParty(cl *client.DamlBindingClient) (string, string) {
	return getAvailableUser(cl), getAvailableParty(cl)
}
