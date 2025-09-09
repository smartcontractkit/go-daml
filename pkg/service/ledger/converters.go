package ledger

import (
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2"
	"github.com/noders-team/go-daml/pkg/model"
)

func commandsToProto(cmd *model.Commands) *v2.Commands {
	if cmd == nil {
		return nil
	}

	pbCmd := &v2.Commands{
		WorkflowId:   cmd.WorkflowID,
		UserId:       cmd.UserID,
		CommandId:    cmd.CommandID,
		Commands:     commandsArrayToProto(cmd.Commands),
		ActAs:        cmd.ActAs,
		ReadAs:       cmd.ReadAs,
		SubmissionId: cmd.SubmissionID,
	}

	if cmd.MinLedgerTimeAbs != nil {
		pbCmd.MinLedgerTimeAbs = timestamppb.New(*cmd.MinLedgerTimeAbs)
	}

	if cmd.MinLedgerTimeRel != nil {
		pbCmd.MinLedgerTimeRel = durationpb.New(*cmd.MinLedgerTimeRel)
	}

	switch dp := cmd.DeduplicationPeriod.(type) {
	case model.DeduplicationDuration:
		pbCmd.DeduplicationPeriod = &v2.Commands_DeduplicationDuration{
			DeduplicationDuration: durationpb.New(dp.Duration),
		}
	case model.DeduplicationOffset:
		pbCmd.DeduplicationPeriod = &v2.Commands_DeduplicationOffset{
			DeduplicationOffset: dp.Offset,
		}
	}

	return pbCmd
}

func commandsArrayToProto(cmds []*model.Command) []*v2.Command {
	if cmds == nil {
		return nil
	}

	result := make([]*v2.Command, len(cmds))
	for i, cmd := range cmds {
		result[i] = commandToProto(cmd)
	}
	return result
}

func commandToProto(cmd *model.Command) *v2.Command {
	if cmd == nil {
		return nil
	}

	pbCmd := &v2.Command{}

	switch c := cmd.Command.(type) {
	case model.CreateCommand:
		pbCmd.Command = &v2.Command_Create{
			Create: &v2.CreateCommand{
				TemplateId: &v2.Identifier{
					PackageId:  "", // TODO: Parse template ID
					ModuleName: "", // TODO: Parse template ID
					EntityName: c.TemplateID,
				},
				CreateArguments: nil, // TODO: Convert arguments to proto Value
			},
		}
	case model.ExerciseCommand:
		pbCmd.Command = &v2.Command_Exercise{
			Exercise: &v2.ExerciseCommand{
				ContractId: c.ContractID,
				TemplateId: &v2.Identifier{
					PackageId:  "", // TODO: Parse template ID
					ModuleName: "", // TODO: Parse template ID
					EntityName: c.TemplateID,
				},
				Choice:         c.Choice,
				ChoiceArgument: nil, // TODO: Convert arguments to proto Value
			},
		}
	case model.ExerciseByKeyCommand:
		pbCmd.Command = &v2.Command_ExerciseByKey{
			ExerciseByKey: &v2.ExerciseByKeyCommand{
				TemplateId: &v2.Identifier{
					PackageId:  "", // TODO: Parse template ID
					ModuleName: "", // TODO: Parse template ID
					EntityName: c.TemplateID,
				},
				ContractKey:    nil, // TODO: Convert key to proto Value
				Choice:         c.Choice,
				ChoiceArgument: nil, // TODO: Convert arguments to proto Value
			},
		}
	}

	return pbCmd
}

func transactionFilterToProto(filter *model.TransactionFilter) *v2.TransactionFilter {
	if filter == nil {
		return nil
	}

	pbFilter := &v2.TransactionFilter{
		FiltersByParty: make(map[string]*v2.Filters),
	}

	for party, filters := range filter.FiltersByParty {
		pbFilter.FiltersByParty[party] = filtersToProto(filters)
	}

	return pbFilter
}

func filtersToProto(filters *model.Filters) *v2.Filters {
	if filters == nil {
		return nil
	}

	pbFilters := &v2.Filters{}

	// Convert template filters to cumulative filters
	if filters.Inclusive != nil {
		for _, tf := range filters.Inclusive.TemplateFilters {
			pbFilters.Cumulative = append(pbFilters.Cumulative, &v2.CumulativeFilter{
				IdentifierFilter: &v2.CumulativeFilter_TemplateFilter{
					TemplateFilter: templateFilterToProto(tf),
				},
			})
		}
	}

	return pbFilters
}

func templateFilterToProto(tf *model.TemplateFilter) *v2.TemplateFilter {
	if tf == nil {
		return nil
	}

	return &v2.TemplateFilter{
		TemplateId: &v2.Identifier{
			// TODO: Parse template ID
			EntityName: tf.TemplateID,
		},
		IncludeCreatedEventBlob: tf.IncludeCreatedEventBlob,
	}
}

func createdEventFromProto(pb *v2.CreatedEvent) *model.CreatedEvent {
	if pb == nil {
		return nil
	}

	event := &model.CreatedEvent{
		ContractID:  pb.ContractId,
		TemplateID:  pb.TemplateId.String(),
		Signatories: pb.Signatories,
		Observers:   pb.Observers,
	}

	// TODO: Convert proto Value to map[string]interface{}
	// event.CreateArguments = valueToMap(pb.CreateArguments)
	// event.ContractKey = valueToMap(pb.ContractKey)

	return event
}

func archivedEventFromProto(pb *v2.ArchivedEvent) *model.ArchivedEvent {
	if pb == nil {
		return nil
	}

	return &model.ArchivedEvent{
		ContractID: pb.ContractId,
		TemplateID: pb.TemplateId.String(),
	}
}
