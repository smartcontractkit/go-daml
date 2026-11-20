package admin

import (
	"testing"

	adminv2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2/admin"
	"github.com/smartcontractkit/go-daml/pkg/model"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRightConversion_AllSevenVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model *model.Right
		proto *adminv2.Right
	}{
		{
			name:  "participant admin",
			model: &model.Right{Type: model.ParticipantAdmin{}},
			proto: &adminv2.Right{Kind: &adminv2.Right_ParticipantAdmin_{
				ParticipantAdmin: &adminv2.Right_ParticipantAdmin{},
			}},
		},
		{
			name:  "can act as",
			model: &model.Right{Type: model.CanActAs{Party: "alice"}},
			proto: &adminv2.Right{Kind: &adminv2.Right_CanActAs_{
				CanActAs: &adminv2.Right_CanActAs{Party: "alice"},
			}},
		},
		{
			name:  "can read as",
			model: &model.Right{Type: model.CanReadAs{Party: "alice"}},
			proto: &adminv2.Right{Kind: &adminv2.Right_CanReadAs_{
				CanReadAs: &adminv2.Right_CanReadAs{Party: "alice"},
			}},
		},
		{
			name:  "identity provider admin",
			model: &model.Right{Type: model.IdentityProviderAdmin{}},
			proto: &adminv2.Right{Kind: &adminv2.Right_IdentityProviderAdmin_{
				IdentityProviderAdmin: &adminv2.Right_IdentityProviderAdmin{},
			}},
		},
		{
			name:  "can read as any party",
			model: &model.Right{Type: model.CanReadAsAnyParty{}},
			proto: &adminv2.Right{Kind: &adminv2.Right_CanReadAsAnyParty_{
				CanReadAsAnyParty: &adminv2.Right_CanReadAsAnyParty{},
			}},
		},
		{
			name:  "can execute as",
			model: &model.Right{Type: model.CanExecuteAs{Party: "alice"}},
			proto: &adminv2.Right{Kind: &adminv2.Right_CanExecuteAs_{
				CanExecuteAs: &adminv2.Right_CanExecuteAs{Party: "alice"},
			}},
		},
		{
			name:  "can execute as any party",
			model: &model.Right{Type: model.CanExecuteAsAnyParty{}},
			proto: &adminv2.Right{Kind: &adminv2.Right_CanExecuteAsAnyParty_{
				CanExecuteAsAnyParty: &adminv2.Right_CanExecuteAsAnyParty{},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted, err := rightsFromProto([]*adminv2.Right{tt.proto})
			require.NoError(t, err)
			require.Equal(t, []*model.Right{tt.model}, converted)

			roundTripped, err := rightToProto(converted[0])
			require.NoError(t, err)
			require.True(t, proto.Equal(tt.proto, roundTripped))
		})
	}
}

func TestRightConversion_RejectsUnknownKinds(t *testing.T) {
	t.Parallel()

	right, err := rightFromProto(&adminv2.Right{})
	require.Nil(t, right)
	require.EqualError(t, err, "unsupported user right kind <nil>")

	pb, err := rightToProto(&model.Right{})
	require.Nil(t, pb)
	require.EqualError(t, err, "unsupported user right type <nil>")
}
