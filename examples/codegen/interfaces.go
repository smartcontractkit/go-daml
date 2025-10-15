package codegen_test

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

const InterfacePackageID = "8f919735d2daa1abb780808ad1fed686fc9229a039dc659ccb04e5fd5d071c90"

type InterfaceTemplate interface {
	CreateCommand() *model.CreateCommand
	GetTemplateID() string
}

// Transferable is a DAML interface
type Transferable interface {

	// Archive executes the Archive choice
	Archive(contractID string) (*model.ExerciseCommand, error)

	// Transfer executes the Transfer choice
	Transfer(contractID string, args Transfer) (*model.ExerciseCommand, error)
}

func interfaceArgsToMap(args interface{}) map[string]interface{} {
	if args == nil {
		return map[string]interface{}{}
	}

	if m, ok := args.(map[string]interface{}); ok {
		return m
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
	return fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Asset")
}

// CreateCommand returns a CreateCommand for this template
func (t Asset) CreateCommand() *model.CreateCommand {
	args := make(map[string]interface{})

	args["owner"] = map[string]interface{}{"_type": "party", "value": string(t.Owner)}

	args["name"] = string(t.Name)

	args["value"] = int64(t.Value)

	return &model.CreateCommand{
		TemplateID: t.GetTemplateID(),
		Arguments:  args,
	}
}

// Choice methods for Asset

// Archive exercises the Archive choice on this Asset contract
func (t Asset) Archive(contractID string) (*model.ExerciseCommand, error) {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Asset"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]interface{}{},
	}, nil
}

// AssetTransfer exercises the AssetTransfer choice on this Asset contract
func (t Asset) AssetTransfer(contractID string, args AssetTransfer) (*model.ExerciseCommand, error) {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Asset"),
		ContractID: contractID,
		Choice:     "AssetTransfer",
		Arguments:  interfaceArgsToMap(args),
	}, nil
}

// Transfer exercises the Transfer choice on this Asset contract
func (t Asset) Transfer(contractID string, args Transfer) (*model.ExerciseCommand, error) {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Asset"),
		ContractID: contractID,
		Choice:     "Transfer",
		Arguments:  interfaceArgsToMap(args),
	}, nil
}

// Verify interface implementations for Asset

var _ Transferable = (*Asset)(nil)

// AssetTransfer is a Record type
type AssetTransfer struct {
	NewOwner PARTY `json:"newOwner"`
}

// Token is a Template type
type Token struct {
	Issuer PARTY   `json:"issuer"`
	Owner  PARTY   `json:"owner"`
	Amount NUMERIC `json:"amount"`
}

// GetTemplateID returns the template ID for this template
func (t Token) GetTemplateID() string {
	return fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Token")
}

// CreateCommand returns a CreateCommand for this template
func (t Token) CreateCommand() *model.CreateCommand {
	args := make(map[string]interface{})

	args["issuer"] = map[string]interface{}{"_type": "party", "value": string(t.Issuer)}

	args["owner"] = map[string]interface{}{"_type": "party", "value": string(t.Owner)}

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
func (t Token) Archive(contractID string) (*model.ExerciseCommand, error) {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Token"),
		ContractID: contractID,
		Choice:     "Archive",
		Arguments:  map[string]interface{}{},
	}, nil
}

// Transfer exercises the Transfer choice on this Token contract
func (t Token) Transfer(contractID string, args Transfer) (*model.ExerciseCommand, error) {
	return &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Interfaces", "Token"),
		ContractID: contractID,
		Choice:     "Transfer",
		Arguments:  interfaceArgsToMap(args),
	}, nil
}

// Verify interface implementations for Token

var _ Transferable = (*Token)(nil)

// Transfer is a Record type
type Transfer struct {
	NewOwner PARTY `json:"newOwner"`
}

// TransferableView is a Record type
type TransferableView struct {
	Owner PARTY `json:"owner"`
}
