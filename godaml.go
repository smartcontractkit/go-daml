package godaml

import "github.com/smartcontractkit/go-daml/pkg/client"

func Client(token string, grpcAddress string) *client.DamlClient {
	return client.NewDamlClient(token, grpcAddress)
}
