package ledger

import (
	"strings"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v2 "github.com/digital-asset/dazl-client/v8/go/api/com/daml/ledger/api/v2"
	"github.com/noders-team/go-daml/pkg/model"
)

func parseTemplateID(templateID string) (packageID, moduleName, entityName string) {
	parts := strings.Split(templateID, ":")
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	} else if len(parts) == 2 {
		return "", parts[0], parts[1]
	}
	return "", "", templateID
}

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
		packageID, moduleName, entityName := parseTemplateID(c.TemplateID)
		pbCmd.Command = &v2.Command_Create{
			Create: &v2.CreateCommand{
				TemplateId: &v2.Identifier{
					PackageId:  packageID,
					ModuleName: moduleName,
					EntityName: entityName,
				},
				CreateArguments: convertToRecord(c.Arguments),
			},
		}
	case model.ExerciseCommand:
		packageID, moduleName, entityName := parseTemplateID(c.TemplateID)
		pbCmd.Command = &v2.Command_Exercise{
			Exercise: &v2.ExerciseCommand{
				ContractId: c.ContractID,
				TemplateId: &v2.Identifier{
					PackageId:  packageID,
					ModuleName: moduleName,
					EntityName: entityName,
				},
				Choice:         c.Choice,
				ChoiceArgument: mapToValue(c.Arguments),
			},
		}
	case model.ExerciseByKeyCommand:
		packageID, moduleName, entityName := parseTemplateID(c.TemplateID)
		pbCmd.Command = &v2.Command_ExerciseByKey{
			ExerciseByKey: &v2.ExerciseByKeyCommand{
				TemplateId: &v2.Identifier{
					PackageId:  packageID,
					ModuleName: moduleName,
					EntityName: entityName,
				},
				ContractKey:    mapToValue(c.Key),
				Choice:         c.Choice,
				ChoiceArgument: mapToValue(c.Arguments),
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

	packageID, moduleName, entityName := parseTemplateID(tf.TemplateID)
	return &v2.TemplateFilter{
		TemplateId: &v2.Identifier{
			PackageId:  packageID,
			ModuleName: moduleName,
			EntityName: entityName,
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

	// Convert proto Values to map[string]interface{}
	if pb.CreateArguments != nil {
		event.CreateArguments = valueFromRecord(pb.CreateArguments)
	}
	if pb.ContractKey != nil {
		if key := valueFromProto(pb.ContractKey); key != nil {
			if m, ok := key.(map[string]interface{}); ok {
				event.ContractKey = m
			}
		}
	}

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

func valueFromProto(pb *v2.Value) interface{} {
	if pb == nil {
		return nil
	}

	switch v := pb.Sum.(type) {
	case *v2.Value_Unit:
		return map[string]interface{}{"_type": "unit"}
	case *v2.Value_Bool:
		return v.Bool
	case *v2.Value_Int64:
		return v.Int64
	case *v2.Value_Text:
		return v.Text
	case *v2.Value_Numeric:
		return v.Numeric
	case *v2.Value_Party:
		return v.Party
	case *v2.Value_ContractId:
		return v.ContractId
	case *v2.Value_Date:
		return v.Date
	case *v2.Value_Timestamp:
		return v.Timestamp
	case *v2.Value_Optional:
		if v.Optional.Value != nil {
			return valueFromProto(v.Optional.Value)
		}
		return nil
	case *v2.Value_List:
		result := make([]interface{}, len(v.List.Elements))
		for i, elem := range v.List.Elements {
			result[i] = valueFromProto(elem)
		}
		return result
	case *v2.Value_Record:
		if v.Record == nil {
			return nil
		}
		record := make(map[string]interface{})
		for _, field := range v.Record.Fields {
			record[field.Label] = valueFromProto(field.Value)
		}
		return record
	case *v2.Value_TextMap:
		if v.TextMap == nil {
			return nil
		}
		result := make(map[string]interface{})
		for _, entry := range v.TextMap.Entries {
			result[entry.Key] = valueFromProto(entry.Value)
		}
		return result
	default:
		return nil
	}
}

func mapToValue(data interface{}) *v2.Value {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case bool:
		return &v2.Value{Sum: &v2.Value_Bool{Bool: v}}
	case int64:
		return &v2.Value{Sum: &v2.Value_Int64{Int64: v}}
	case int:
		return &v2.Value{Sum: &v2.Value_Int64{Int64: int64(v)}}
	case string:
		return &v2.Value{Sum: &v2.Value_Text{Text: v}}
	case []interface{}:
		elements := make([]*v2.Value, len(v))
		for i, elem := range v {
			elements[i] = mapToValue(elem)
		}
		return &v2.Value{
			Sum: &v2.Value_List{
				List: &v2.List{Elements: elements},
			},
		}
	case map[string]interface{}:
		if typeStr, ok := v["_type"].(string); ok && typeStr == "unit" {
			return &v2.Value{Sum: &v2.Value_Unit{Unit: &emptypb.Empty{}}}
		}
		fields := make([]*v2.RecordField, 0, len(v))
		for key, val := range v {
			if key != "_type" {
				fields = append(fields, &v2.RecordField{
					Label: key,
					Value: mapToValue(val),
				})
			}
		}
		return &v2.Value{
			Sum: &v2.Value_Record{
				Record: &v2.Record{Fields: fields},
			},
		}
	default:
		return nil
	}
}

func convertToRecord(data map[string]interface{}) *v2.Record {
	if data == nil {
		return nil
	}

	fields := make([]*v2.RecordField, 0, len(data))
	for key, val := range data {
		fields = append(fields, &v2.RecordField{
			Label: key,
			Value: mapToValue(val),
		})
	}

	return &v2.Record{Fields: fields}
}

func valueFromRecord(record *v2.Record) map[string]interface{} {
	if record == nil {
		return nil
	}

	result := make(map[string]interface{})
	for _, field := range record.Fields {
		result[field.Label] = valueFromProto(field.Value)
	}
	return result
}
