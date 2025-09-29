package {{.Package}}

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"errors"
	"time"
	
	"github.com/noders-team/go-daml/pkg/model"
)

var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
)

const PackageID = "{{.PackageID}}"

type (
	PARTY string
	TEXT string
	INT64 int64
	BOOL bool
	DECIMAL *big.Int
	NUMERIC *big.Int
	DATE time.Time
	TIMESTAMP time.Time
	UNIT struct{}
	LIST []string
	MAP map[string]interface{}
	OPTIONAL *interface{}
	GENMAP map[string]interface{}
)

// argsToMap converts typed arguments to map for ExerciseCommand
func argsToMap(args interface{}) map[string]interface{} {
	// For now, we'll use a simple approach
	// In practice, you might want to implement proper struct-to-map conversion
	if args == nil {
		return map[string]interface{}{}
	}
	
	// If args is already a map, return it directly
	if m, ok := args.(map[string]interface{}); ok {
		return m
	}
	
	// For structs, you would typically use reflection or JSON marshaling
	// For simplicity, we'll return the args in a generic wrapper
	return map[string]interface{}{
		"args": args,
	}
}


{{$structs := .Structs}}
{{range $structs}}
	{{if eq .RawType "Variant"}}
	// {{capitalise .Name}} is a variant/union type
	type {{capitalise .Name}} struct {
		{{range $field := .Fields}}
		{{capitalise $field.Name}} *{{$field.Type}} `json:"{{$field.Name}},omitempty"`{{end}}
	}

	// MarshalJSON implements custom JSON marshaling for {{capitalise .Name}}
	func (v {{capitalise .Name}}) MarshalJSON() ([]byte, error) {
		{{range $field := .Fields}}
		if v.{{capitalise $field.Name}} != nil {
			return json.Marshal(map[string]interface{}{
				"tag":   "{{$field.Name}}",
				"value": v.{{capitalise $field.Name}},
			})
		}
		{{end}}
		return json.Marshal(map[string]interface{}{})
	}

	// UnmarshalJSON implements custom JSON unmarshaling for {{capitalise .Name}}
	func (v *{{capitalise .Name}}) UnmarshalJSON(data []byte) error {
		var tagged struct {
			Tag   string          `json:"tag"`
			Value json.RawMessage `json:"value"`
		}
		
		if err := json.Unmarshal(data, &tagged); err != nil {
			return err
		}
		
		switch tagged.Tag {
		{{range $field := .Fields}}
		case "{{$field.Name}}":
			var value {{$field.Type}}
			if err := json.Unmarshal(tagged.Value, &value); err != nil {
				return err
			}
			v.{{capitalise $field.Name}} = &value
		{{end}}
		default:
			return fmt.Errorf("unknown tag: %s", tagged.Tag)
		}
		
		return nil
	}
	{{else if eq .RawType "Enum"}}
	// {{capitalise .Name}} is an enum type
	type {{capitalise .Name}} string

	const (
		{{$structName := .Name}}{{range $field := .Fields}}
		{{capitalise $structName}}{{$field.Name}} {{capitalise $structName}} = "{{$field.Name}}"{{end}}
	)
	{{else}}
	// {{capitalise .Name}} is a {{.RawType}} type
	type {{capitalise .Name}} struct {
		{{range $field := .Fields}}
		{{capitalise $field.Name}} {{$field.Type}} `json:"{{$field.Name}}"`{{end}}
	}
	{{if and .IsTemplate .Key}}
	
	// GetKey returns the key for this template as a string
	func (t {{capitalise .Name}}) GetKey() string {
		{{if eq .Key.Type "TEXT"}}
		return string(t.{{capitalise .Key.Name}})
		{{else if eq .Key.Type "PARTY"}}
		return string(t.{{capitalise .Key.Name}})
		{{else if eq .Key.Type "INT64"}}
		return fmt.Sprintf("%d", t.{{capitalise .Key.Name}})
		{{else}}
		return fmt.Sprintf("%v", t.{{capitalise .Key.Name}})
		{{end}}
	}
	{{end}}
	{{if and .IsTemplate .Choices}}
	{{$templateName := .Name}}
	{{$moduleName := .ModuleName}}
	// Choice methods for {{capitalise .Name}}
	{{range $choice := .Choices}}
	// {{capitalise $choice.Name}} exercises the {{$choice.Name}} choice on this {{capitalise $templateName}} contract
	func (t {{capitalise $templateName}}) {{capitalise $choice.Name}}(contractID string{{if $choice.ArgType}}, args {{$choice.ArgType}}{{end}}) *model.ExerciseCommand {
		return &model.ExerciseCommand{
			TemplateID: fmt.Sprintf("%s:%s:%s", PackageID, "{{$moduleName}}", "{{capitalise $templateName}}"),
			ContractID: contractID,
			Choice: "{{$choice.Name}}",
			{{if $choice.ArgType}}Arguments: argsToMap(args),{{else}}Arguments: map[string]interface{}{},{{end}}
		}
	}
	{{end}}
	{{end}}
	{{end}}
{{end}}
