package admin

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	adminv2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2/admin"
	"github.com/smartcontractkit/go-daml/pkg/model"
)

type UserManagement interface {
	CreateUser(ctx context.Context, user *model.User, rights []*model.Right) (*model.User, error)
	GetUser(ctx context.Context, userID string) (*model.User, error)
	DeleteUser(ctx context.Context, userID string) error
	GrantUserRights(ctx context.Context, userID, identityProviderID string, rights []*model.Right) ([]*model.Right, error)
	RevokeUserRights(ctx context.Context, userID string, rights []*model.Right) ([]*model.Right, error)
	ListUserRights(ctx context.Context, userID string) ([]*model.Right, error)
	ListUsers(ctx context.Context) ([]*model.User, error)
}

type userManagement struct {
	client adminv2.UserManagementServiceClient
}

func NewUserManagementClient(conn *grpc.ClientConn) *userManagement {
	client := adminv2.NewUserManagementServiceClient(conn)
	return &userManagement{
		client: client,
	}
}

func (c *userManagement) CreateUser(ctx context.Context, user *model.User, rights []*model.Right) (*model.User, error) {
	protoRights, err := rightsToProto(rights)
	if err != nil {
		return nil, err
	}
	request := &adminv2.CreateUserRequest{
		User:   userToProto(user),
		Rights: protoRights,
	}

	resp, err := c.client.CreateUser(ctx, request)
	if err != nil {
		return nil, err
	}

	return userFromProto(resp.GetUser()), nil
}

func (c *userManagement) GetUser(ctx context.Context, userID string) (*model.User, error) {
	req := &adminv2.GetUserRequest{
		UserId: userID,
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return userFromProto(resp.User), nil
}

func (c *userManagement) ListUsers(ctx context.Context) ([]*model.User, error) {
	req := &adminv2.ListUsersRequest{}

	resp, err := c.client.ListUsers(ctx, req)
	if err != nil {
		return nil, err
	}

	return usersFromProto(resp.Users), nil
}

func (c *userManagement) DeleteUser(ctx context.Context, userID string) error {
	req := &adminv2.DeleteUserRequest{
		UserId: userID,
	}

	_, err := c.client.DeleteUser(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *userManagement) GrantUserRights(ctx context.Context, userID, identityProviderID string, rights []*model.Right) ([]*model.Right, error) {
	protoRights, err := rightsToProto(rights)
	if err != nil {
		return nil, err
	}
	req := &adminv2.GrantUserRightsRequest{
		UserId:             userID,
		IdentityProviderId: identityProviderID,
		Rights:             protoRights,
	}

	resp, err := c.client.GrantUserRights(ctx, req)
	if err != nil {
		return nil, err
	}

	return rightsFromProto(resp.NewlyGrantedRights)
}

func (c *userManagement) RevokeUserRights(ctx context.Context, userID string, rights []*model.Right) ([]*model.Right, error) {
	protoRights, err := rightsToProto(rights)
	if err != nil {
		return nil, err
	}
	req := &adminv2.RevokeUserRightsRequest{
		UserId: userID,
		Rights: protoRights,
	}

	resp, err := c.client.RevokeUserRights(ctx, req)
	if err != nil {
		return nil, err
	}

	return rightsFromProto(resp.NewlyRevokedRights)
}

func (c *userManagement) ListUserRights(ctx context.Context, userID string) ([]*model.Right, error) {
	req := &adminv2.ListUserRightsRequest{
		UserId: userID,
	}

	resp, err := c.client.ListUserRights(ctx, req)
	if err != nil {
		return nil, err
	}

	return rightsFromProto(resp.Rights)
}

func userFromProto(pb *adminv2.User) *model.User {
	if pb == nil {
		return nil
	}
	metadata := make(map[string]string)
	if pb.Metadata != nil {
		metadata = pb.Metadata.Annotations
	}
	return &model.User{
		ID:                 pb.Id,
		PrimaryParty:       pb.PrimaryParty,
		IsDeactivated:      pb.IsDeactivated,
		Metadata:           metadata,
		IdentityProviderID: pb.IdentityProviderId,
	}
}

func userToProto(u *model.User) *adminv2.User {
	if u == nil {
		return nil
	}
	var metadata *adminv2.ObjectMeta
	if len(u.Metadata) > 0 {
		metadata = &adminv2.ObjectMeta{
			Annotations: u.Metadata,
		}
	}
	return &adminv2.User{
		Id:                 u.ID,
		PrimaryParty:       u.PrimaryParty,
		IsDeactivated:      u.IsDeactivated,
		Metadata:           metadata,
		IdentityProviderId: u.IdentityProviderID,
	}
}

func rightFromProto(pb *adminv2.Right) (*model.Right, error) {
	if pb == nil {
		return nil, nil
	}
	r := &model.Right{}
	switch rt := pb.Kind.(type) {
	case *adminv2.Right_CanActAs_:
		r.Type = model.CanActAs{Party: rt.CanActAs.GetParty()}
	case *adminv2.Right_CanReadAs_:
		r.Type = model.CanReadAs{Party: rt.CanReadAs.GetParty()}
	case *adminv2.Right_ParticipantAdmin_:
		r.Type = model.ParticipantAdmin{}
	case *adminv2.Right_IdentityProviderAdmin_:
		r.Type = model.IdentityProviderAdmin{}
	case *adminv2.Right_CanReadAsAnyParty_:
		r.Type = model.CanReadAsAnyParty{}
	case *adminv2.Right_CanExecuteAs_:
		r.Type = model.CanExecuteAs{Party: rt.CanExecuteAs.GetParty()}
	case *adminv2.Right_CanExecuteAsAnyParty_:
		r.Type = model.CanExecuteAsAnyParty{}
	default:
		return nil, fmt.Errorf("unsupported user right kind %T", pb.Kind)
	}
	return r, nil
}

func rightToProto(r *model.Right) (*adminv2.Right, error) {
	if r == nil {
		return nil, nil
	}
	pb := &adminv2.Right{}
	switch rt := r.Type.(type) {
	case model.CanActAs:
		pb.Kind = &adminv2.Right_CanActAs_{
			CanActAs: &adminv2.Right_CanActAs{Party: rt.Party},
		}
	case model.CanReadAs:
		pb.Kind = &adminv2.Right_CanReadAs_{
			CanReadAs: &adminv2.Right_CanReadAs{Party: rt.Party},
		}
	case model.ParticipantAdmin:
		pb.Kind = &adminv2.Right_ParticipantAdmin_{
			ParticipantAdmin: &adminv2.Right_ParticipantAdmin{},
		}
	case model.IdentityProviderAdmin:
		pb.Kind = &adminv2.Right_IdentityProviderAdmin_{
			IdentityProviderAdmin: &adminv2.Right_IdentityProviderAdmin{},
		}
	case model.CanReadAsAnyParty:
		pb.Kind = &adminv2.Right_CanReadAsAnyParty_{
			CanReadAsAnyParty: &adminv2.Right_CanReadAsAnyParty{},
		}
	case model.CanExecuteAs:
		pb.Kind = &adminv2.Right_CanExecuteAs_{
			CanExecuteAs: &adminv2.Right_CanExecuteAs{Party: rt.Party},
		}
	case model.CanExecuteAsAnyParty:
		pb.Kind = &adminv2.Right_CanExecuteAsAnyParty_{
			CanExecuteAsAnyParty: &adminv2.Right_CanExecuteAsAnyParty{},
		}
	default:
		return nil, fmt.Errorf("unsupported user right type %T", r.Type)
	}
	return pb, nil
}

func rightsFromProto(pbs []*adminv2.Right) ([]*model.Right, error) {
	rights := make([]*model.Right, len(pbs))
	for i, pb := range pbs {
		right, err := rightFromProto(pb)
		if err != nil {
			return nil, fmt.Errorf("convert user right at index %d: %w", i, err)
		}
		rights[i] = right
	}
	return rights, nil
}

func rightsToProto(rights []*model.Right) ([]*adminv2.Right, error) {
	pbs := make([]*adminv2.Right, len(rights))
	for i, r := range rights {
		pb, err := rightToProto(r)
		if err != nil {
			return nil, fmt.Errorf("convert user right at index %d: %w", i, err)
		}
		pbs[i] = pb
	}
	return pbs, nil
}

func usersFromProto(pbs []*adminv2.User) []*model.User {
	users := make([]*model.User, len(pbs))
	for i, pb := range pbs {
		users[i] = userFromProto(pb)
	}
	return users
}
