package model

import (
	"time"
)

type Commands struct {
	WorkflowID          string
	UserID              string
	CommandID           string
	Commands            []*Command
	DeduplicationPeriod DeduplicationPeriod
	MinLedgerTimeAbs    *time.Time
	MinLedgerTimeRel    *time.Duration
	ActAs               []string
	ReadAs              []string
	SubmissionID        string
}

type DeduplicationPeriod interface {
	isDeduplicationPeriod()
}

type DeduplicationDuration struct {
	Duration time.Duration
}

func (DeduplicationDuration) isDeduplicationPeriod() {}

type DeduplicationOffset struct {
	Offset int64
}

func (DeduplicationOffset) isDeduplicationPeriod() {}

type Command struct {
	Command CommandType
}

type CommandType interface {
	isCommandType()
}

type CreateCommand struct {
	TemplateID string
	Arguments  map[string]interface{}
}

func (CreateCommand) isCommandType() {}

type ExerciseCommand struct {
	ContractID string
	TemplateID string
	Choice     string
	Arguments  map[string]interface{}
}

func (ExerciseCommand) isCommandType() {}

type ExerciseByKeyCommand struct {
	TemplateID string
	Key        map[string]interface{}
	Choice     string
	Arguments  map[string]interface{}
}

func (ExerciseByKeyCommand) isCommandType() {}

type CompletionStreamRequest struct {
	UserID         string
	Parties        []string
	BeginExclusive int64
}

type CompletionStreamResponse struct {
	Response CompletionResponse
}

type CompletionResponse interface {
	isCompletionResponse()
}

type Completion struct {
	CommandID    string
	Status       Status
	UpdateID     string
	TransactionID string
	SubmissionID string
	CompletedAt  *time.Time
	Offset       int64
}

func (Completion) isCompletionResponse() {}

type OffsetCheckpoint struct {
	Offset int64
}

func (OffsetCheckpoint) isCompletionResponse() {}

type Status interface {
	isStatus()
}

type StatusOK struct{}

func (StatusOK) isStatus() {}

type StatusError struct {
	Code    int32
	Message string
}

func (StatusError) isStatus() {}

type SubmitRequest struct {
	Commands *Commands
}

type SubmitResponse struct{}

type SubmitAndWaitRequest struct {
	Commands *Commands
}

type SubmitAndWaitResponse struct {
	UpdateID         string
	CompletionOffset int64
}