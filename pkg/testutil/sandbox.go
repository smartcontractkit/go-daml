package testutil

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/freeport"
	"github.com/smartcontractkit/go-daml/pkg/model"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/smartcontractkit/go-daml/pkg/client"
)

const (
	DamlSandboxImage   = "digitalasset/daml-sdk"
	DamlSandboxVersion = "3.5.0-snapshot.20251106.0"
	SandboxUserId      = "app-provider"
)

type Output struct {
	Container     testcontainers.Container
	BindingClient *client.DamlBindingClient
}

func CreateSandbox(t *testing.T) (*Output, error) {
	t.Helper()

	// Allocate two ports
	ports := freeport.GetN(t, 2)
	exposedPorts := []string{
		fmt.Sprintf("%d:%d", ports[0], 6865),
		fmt.Sprintf("%d:%d", ports[1], 6866),
	}

	log.Info().Msg("Starting sandbox container...")
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", DamlSandboxImage, DamlSandboxVersion),
		ExposedPorts: exposedPorts,
		WaitingFor:   wait.ForLog("Canton sandbox is ready.").WithStartupTimeout(time.Minute * 10),
		Files: []testcontainers.ContainerFile{
			{
				Reader:            strings.NewReader(GetCantonConfig()),
				ContainerFilePath: "/canton/canton.conf",
				FileMode:          0755,
			},
		},
		ImagePlatform: "linux/amd64",
		Cmd: []string{
			"daml",
			"sandbox",
			"-c", "/canton/canton.conf",
			"--debug",
		},
	}
	container, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	grpcAddress := fmt.Sprintf("localhost:%d", ports[0])
	adminAddress := fmt.Sprintf("localhost:%d", ports[1])

	c, err := client.NewDamlClient("", grpcAddress).WithAdminAddress(adminAddress).Build(t.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to build daml client: %w", err)
	}

	users, err := c.UserMng.ListUsers(t.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	userExists := false
	for _, u := range users {
		if u.ID == SandboxUserId {
			userExists = true
		}
	}

	if !userExists {
		partyDetails, err := c.PartyMng.AllocateParty(t.Context(), "", nil, "")
		if err != nil {
			return nil, fmt.Errorf("failed to allocate party: %w", err)
		}
		log.Debug().Str("PartyId", partyDetails.Party).Msg("Allocated party")
		user := &model.User{
			ID:           SandboxUserId,
			PrimaryParty: partyDetails.Party,
		}
		rights := []*model.Right{
			{Type: model.CanActAs{Party: partyDetails.Party}},
			{Type: model.CanReadAs{Party: partyDetails.Party}},
		}
		_, err = c.UserMng.CreateUser(t.Context(), user, rights)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	log.Info().Msg("Sandbox ready")

	return &Output{
		Container:     container,
		BindingClient: c,
	}, nil
}

func GetCantonConfig() string {
	//language=HOCON
	return `
canton {
  mediators {
    mediator1 {
      admin-api.port = 6869
    }
  }
  sequencers {
    sequencer1 {
      admin-api.port = 6868
      public-api.port = 6867
      sequencer {
        type = reference
        config.storage.type = memory
      }
      storage.type = memory
    }
  }
  participants {
    sandbox {
      storage.type = memory
      admin-api {
	    address = "0.0.0.0"
	  	port = 6866
	  }
      ledger-api {
        address = "0.0.0.0"
        port = 6865
        user-management-service.enabled = true
      }
    }
  }
}
	`
}
