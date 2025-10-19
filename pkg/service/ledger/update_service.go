package ledger

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc"

	v2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2"
	"github.com/noders-team/go-daml/pkg/model"
)

type UpdateService interface {
	GetUpdates(ctx context.Context, req *model.GetUpdatesRequest) (<-chan *model.GetUpdatesResponse, <-chan error)
	GetTransactionByID(ctx context.Context, req *model.GetTransactionByIDRequest) (*model.GetTransactionResponse, error)
	GetTransactionByOffset(ctx context.Context, req *model.GetTransactionByOffsetRequest) (*model.GetTransactionResponse, error)
}

type GetTransactionResponseTyped[T any] struct {
	Transaction *TransactionTyped[T]
}

type TransactionTyped[T any] struct {
	UpdateID    string
	CommandID   string
	WorkflowID  string
	EffectiveAt *time.Time
	Events      []*EventTyped[T]
	Offset      int64
}

type EventTyped[T any] struct {
	Created   *CreatedEventTyped[T]
	Archived  *model.ArchivedEvent
	Exercised *model.ExercisedEvent
}

type CreatedEventTyped[T any] struct {
	Offset           int64
	NodeID           int32
	ContractID       string
	TemplateID       string
	ContractKey      map[string]interface{}
	CreateArguments  *T
	CreatedEventBlob []byte
	InterfaceViews   []*model.InterfaceView
	WitnessParties   []string
	Signatories      []string
	Observers        []string
	CreatedAt        *time.Time
	PackageName      string
}

type updateService struct {
	client v2.UpdateServiceClient
}

func NewUpdateServiceClient(conn *grpc.ClientConn) *updateService {
	client := v2.NewUpdateServiceClient(conn)
	return &updateService{
		client: client,
	}
}

func (c *updateService) Client() v2.UpdateServiceClient {
	return c.client
}

func (c *updateService) GetUpdates(ctx context.Context, req *model.GetUpdatesRequest) (<-chan *model.GetUpdatesResponse, <-chan error) {
	protoReq := &v2.GetUpdatesRequest{
		BeginExclusive: req.BeginExclusive,
		Filter:         transactionFilterToProto(req.Filter),
		Verbose:        req.Verbose,
	}

	if req.EndInclusive != nil {
		protoReq.EndInclusive = req.EndInclusive
	}

	stream, err := c.client.GetUpdates(ctx, protoReq)
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- err
		close(errCh)
		return nil, errCh
	}

	responseCh := make(chan *model.GetUpdatesResponse)
	errCh := make(chan error, 1)

	go func() {
		defer close(responseCh)
		defer close(errCh)

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errCh <- err
				return
			}

			modelResp := getUpdatesResponseFromProto(resp)
			if modelResp != nil {
				select {
				case responseCh <- modelResp:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return responseCh, errCh
}

func (c *updateService) GetTransactionByID(ctx context.Context, req *model.GetTransactionByIDRequest) (*model.GetTransactionResponse, error) {
	protoReq := &v2.GetTransactionByIdRequest{
		UpdateId:          req.UpdateID,
		RequestingParties: req.RequestingParties,
	}

	resp, err := c.client.GetTransactionById(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return getTransactionResponseFromProto(resp), nil
}

func GetTransactionByIDTyped[T any](ctx context.Context, client v2.UpdateServiceClient, req *model.GetTransactionByIDRequest) (*GetTransactionResponseTyped[T], error) {
	protoReq := &v2.GetTransactionByIdRequest{
		UpdateId:          req.UpdateID,
		RequestingParties: req.RequestingParties,
	}

	resp, err := client.GetTransactionById(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return getTransactionResponseTypedFromProto[T](resp)
}

func (c *updateService) GetTransactionByOffset(ctx context.Context, req *model.GetTransactionByOffsetRequest) (*model.GetTransactionResponse, error) {
	protoReq := &v2.GetTransactionByOffsetRequest{
		Offset:            req.Offset,
		RequestingParties: req.RequestingParties,
	}

	resp, err := c.client.GetTransactionByOffset(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return getTransactionResponseFromProto(resp), nil
}

func getUpdatesResponseFromProto(pb *v2.GetUpdatesResponse) *model.GetUpdatesResponse {
	if pb == nil {
		return nil
	}

	resp := &model.GetUpdatesResponse{
		Update: &model.Update{},
	}

	switch update := pb.Update.(type) {
	case *v2.GetUpdatesResponse_Transaction:
		if update.Transaction != nil {
			resp.Update.Transaction = transactionFromProto(update.Transaction)
		}
	case *v2.GetUpdatesResponse_Reassignment:
		if update.Reassignment != nil {
			resp.Update.Reassignment = reassignmentFromProto(update.Reassignment)
		}
	case *v2.GetUpdatesResponse_OffsetCheckpoint:
		resp.Update.OffsetCheckpoint = &model.OffsetCheckpoint{
			Offset: update.OffsetCheckpoint.Offset,
		}
	}

	return resp
}

func getTransactionResponseFromProto(pb *v2.GetTransactionResponse) *model.GetTransactionResponse {
	if pb == nil {
		return nil
	}

	return &model.GetTransactionResponse{
		Transaction: transactionFromProto(pb.Transaction),
	}
}

func transactionFromProto(pb *v2.Transaction) *model.Transaction {
	if pb == nil {
		return nil
	}

	tx := &model.Transaction{
		UpdateID:   pb.UpdateId,
		CommandID:  pb.CommandId,
		WorkflowID: pb.WorkflowId,
		Offset:     pb.Offset,
	}

	if pb.EffectiveAt != nil {
		t := pb.EffectiveAt.AsTime()
		tx.EffectiveAt = &t
	}

	for _, event := range pb.Events {
		tx.Events = append(tx.Events, eventFromProto(event))
	}

	return tx
}

func eventFromProto(pb *v2.Event) *model.Event {
	if pb == nil {
		return nil
	}

	event := &model.Event{}

	switch e := pb.Event.(type) {
	case *v2.Event_Created:
		event.Created = createdEventFromProto(e.Created)
	case *v2.Event_Archived:
		event.Archived = archivedEventFromProto(e.Archived)
	case *v2.Event_Exercised:
		event.Exercised = exercisedEventFromProto(e.Exercised)
	}

	return event
}

func reassignmentFromProto(pb *v2.Reassignment) *model.Reassignment {
	if pb == nil {
		return nil
	}

	r := &model.Reassignment{
		UpdateID: pb.UpdateId,
		Offset:   pb.Offset,
	}

	if pb.RecordTime != nil {
		t := pb.RecordTime.AsTime()
		r.SubmittedAt = &t
	}

	for _, event := range pb.Events {
		switch e := event.Event.(type) {
		case *v2.ReassignmentEvent_Unassigned:
			if e.Unassigned != nil {
				r.UnassignID = e.Unassigned.UnassignId
				r.Source = e.Unassigned.Source
				r.Target = e.Unassigned.Target
				r.Counter = int64(e.Unassigned.ReassignmentCounter)
				if e.Unassigned.AssignmentExclusivity != nil {
					t := e.Unassigned.AssignmentExclusivity.AsTime()
					r.Unassigned = &t
				}
			}
		case *v2.ReassignmentEvent_Assigned:
			if e.Assigned != nil {
				if r.UnassignID == "" {
					r.UnassignID = e.Assigned.UnassignId
				}
				if r.Source == "" {
					r.Source = e.Assigned.Source
				}
				if r.Target == "" {
					r.Target = e.Assigned.Target
				}
				if r.Counter == 0 {
					r.Counter = int64(e.Assigned.ReassignmentCounter)
				}
				r.Reassigned = r.SubmittedAt
			}
		}
	}

	return r
}

func getTransactionResponseTypedFromProto[T any](pb *v2.GetTransactionResponse) (*GetTransactionResponseTyped[T], error) {
	if pb == nil {
		return nil, nil
	}

	tx, err := transactionTypedFromProto[T](pb.Transaction)
	if err != nil {
		return nil, err
	}

	return &GetTransactionResponseTyped[T]{
		Transaction: tx,
	}, nil
}

func transactionTypedFromProto[T any](pb *v2.Transaction) (*TransactionTyped[T], error) {
	if pb == nil {
		return nil, nil
	}

	tx := &TransactionTyped[T]{
		UpdateID:   pb.UpdateId,
		CommandID:  pb.CommandId,
		WorkflowID: pb.WorkflowId,
		Offset:     pb.Offset,
	}

	if pb.EffectiveAt != nil {
		t := pb.EffectiveAt.AsTime()
		tx.EffectiveAt = &t
	}

	for _, event := range pb.Events {
		typedEvent, err := eventTypedFromProto[T](event)
		if err != nil {
			return nil, err
		}
		tx.Events = append(tx.Events, typedEvent)
	}

	return tx, nil
}

func eventTypedFromProto[T any](pb *v2.Event) (*EventTyped[T], error) {
	if pb == nil {
		return nil, nil
	}

	event := &EventTyped[T]{}

	switch e := pb.Event.(type) {
	case *v2.Event_Created:
		typedCreated, err := createdEventTypedFromProto[T](e.Created)
		if err != nil {
			return nil, err
		}
		event.Created = typedCreated
	case *v2.Event_Archived:
		event.Archived = archivedEventFromProto(e.Archived)
	case *v2.Event_Exercised:
		event.Exercised = exercisedEventFromProto(e.Exercised)
	}

	return event, nil
}

func createdEventTypedFromProto[T any](pb *v2.CreatedEvent) (*CreatedEventTyped[T], error) {
	if pb == nil {
		return nil, nil
	}

	event := &CreatedEventTyped[T]{
		Offset:           pb.Offset,
		NodeID:           pb.NodeId,
		ContractID:       pb.ContractId,
		CreatedEventBlob: pb.CreatedEventBlob,
		WitnessParties:   pb.WitnessParties,
		Signatories:      pb.Signatories,
		Observers:        pb.Observers,
		PackageName:      pb.PackageName,
	}

	if pb.TemplateId != nil {
		event.TemplateID = identifierToString(pb.TemplateId)
	}

	if pb.CreateArguments != nil {
		var createArgs T
		if err := recordToStruct(pb.CreateArguments, &createArgs); err != nil {
			return nil, err
		}
		event.CreateArguments = &createArgs
	}

	if pb.ContractKey != nil {
		if key := valueFromProto(pb.ContractKey); key != nil {
			if m, ok := key.(map[string]interface{}); ok {
				event.ContractKey = m
			}
		}
	}

	if pb.CreatedAt != nil {
		t := pb.CreatedAt.AsTime()
		event.CreatedAt = &t
	}

	if len(pb.InterfaceViews) > 0 {
		event.InterfaceViews = make([]*model.InterfaceView, len(pb.InterfaceViews))
		for i, iv := range pb.InterfaceViews {
			event.InterfaceViews[i] = interfaceViewFromProto(iv)
		}
	}

	return event, nil
}
