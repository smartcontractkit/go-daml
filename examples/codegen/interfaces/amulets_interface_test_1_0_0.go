package interfaces_test

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/noders-team/go-daml/pkg/model"
	. "github.com/noders-team/go-daml/pkg/types"
)

var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
)

const PackageID = "8f919735d2daa1abb780808ad1fed686fc9229a039dc659ccb04e5fd5d071c90"

type Template interface {
	CreateCommand() *model.CreateCommand
	GetTemplateID() string
}

// Transferable is a DAML interface
type Transferable interface {
	// Archive executes the Archive choice
	Archive(contractID string) *model.ExerciseCommand

	// Transfer executes the Transfer choice
	Transfer(contractID string, args Transfer) *model.ExerciseCommand
}

func argsToMap(args interface{}) map[string]interface{} {
	if args == nil {
		return map[string]interface{}{}
	}

	if m, ok := args.(map[string]interface{}); ok {
		return m
	}

	// Check if the type has a toMap method
	type mapper interface {
		toMap() map[string]interface{}
	}

	if mapper, ok := args.(mapper); ok {
		return mapper.toMap()
	}

	return map[string]interface{}{
		"args": args,
	}
}

// Asset is a Template type
type Asset struct {
	Owner PARTY `json:"owner"`
	Name  TEXT  `json:"name"`
	Value INT64 `json:"value"`
}

// GetTemplateID returns the template ID for this template
func (t Asset) GetTemplateID() string {
	return fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Asset")
}

// CreateCommand returns a CreateCommand for this template
func (t Asset) CreateCommand() *model.CreateCommand {
	args := make(map[string]interface{})

	args["owner"] = t.Owner.ToMap()

	args["name"] = string(t.Name)

	args["value"] = int64(t.Value)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// Choice methods for Asset

// Archive exercises the Archive choice on this Asset contract
func (t Asset) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Asset"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]interface{}{},
	}
}

// AssetTransfer exercises the AssetTransfer choice on this Asset contract
func (t Asset) AssetTransfer(contractID string, args AssetTransfer) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Asset"),
		ContractID: contractID,
		Choice:     "AssetTransfer",
		Arguments:  argsToMap(args),
	}
}

// Transfer exercises the Transfer choice on this Asset contract via the Transferable interface
func (t Asset) Transfer(contractID string, args Transfer) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Transferable"),
		ContractID: contractID,
		Choice:     "Transfer",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for Asset

var _ Transferable = (*Asset)(nil)

// AssetTransfer is a Record type
type AssetTransfer struct {
	NewOwner PARTY `json:"newOwner"`
}

// toMap converts AssetTransfer to a map for DAML arguments
func (t AssetTransfer) toMap() map[string]interface{} {
	return map[string]interface{}{
		"newOwner": t.NewOwner.ToMap(),
	}
}

// Token is a Template type
type Token struct {
	Issuer PARTY   `json:"issuer"`
	Owner  PARTY   `json:"owner"`
	Amount NUMERIC `json:"amount"`
}

// GetTemplateID returns the template ID for this template
func (t Token) GetTemplateID() string {
	return fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Token")
}

// CreateCommand returns a CreateCommand for this template
func (t Token) CreateCommand() *model.CreateCommand {
	args := make(map[string]interface{})

	args["issuer"] = t.Issuer.ToMap()

	args["owner"] = t.Owner.ToMap()

	if t.Amount != nil {
		args["amount"] = (*big.Int)(t.Amount)
	}

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// Choice methods for Token

// Archive exercises the Archive choice on this Token contract
func (t Token) Archive(contractID string) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Token"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]interface{}{},
	}
}

// Transfer exercises the Transfer choice on this Token contract via the Transferable interface
func (t Token) Transfer(contractID string, args Transfer) *model.ExerciseCommand {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "Interfaces", "Transferable"),
		ContractID: contractID,
		Choice:     "Transfer",
		Arguments:  argsToMap(args),
	}
}

// Verify interface implementations for Token

var _ Transferable = (*Token)(nil)

// Transfer is a Record type
type Transfer struct {
	NewOwner PARTY `json:"newOwner"`
}

// toMap converts Transfer to a map for DAML arguments
func (t Transfer) toMap() map[string]interface{} {
	return map[string]interface{}{
		"newOwner": t.NewOwner.ToMap(),
	}
}

// TransferableView is a Record type
type TransferableView struct {
	Owner PARTY `json:"owner"`
}

// toMap converts TransferableView to a map for DAML arguments
func (t TransferableView) toMap() map[string]interface{} {
	return map[string]interface{}{
		"owner": t.Owner.ToMap(),
	}
}
