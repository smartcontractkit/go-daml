package testutil

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/freeport"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/smartcontractkit/go-daml/pkg/client"
)

const (
	DamlSandboxImage   = "digitalasset/daml-sdk"
	DamlSandboxVersion = "3.5.0-snapshot.20251106.0"
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

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", DamlSandboxImage, DamlSandboxVersion),
		ExposedPorts: exposedPorts,
		WaitingFor:   wait.ForLog("Canton sandbox is ready.").WithStartupTimeout(time.Minute * 5),
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

	for _, user := range users {
		log.Info().Str("UserID", user.ID).Str("PrimaryParty", user.PrimaryParty).Msg("Existing user")
	}

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
