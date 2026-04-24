package codegen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"

	"github.com/smartcontractkit/go-daml/codegen/model"
)

type tmplData struct {
	Package            string
	PackageID          string
	PackageName        string
	SdkVersion         string
	Structs            map[string]*model.TmplStruct
	IsMainDalf         bool
	GenerateHexCodec   bool
	ChoiceArgTypes     map[string]bool // Types used as choice arguments (for Encode functions)
	ChoiceArgChoices   map[string]string
	ParamEncoderNames  map[string]bool
	EncoderMethodNames map[string]string
	ImportedPackages   []model.ExternalPackage
}

//go:embed source.go.tmpl
var tmplSource string

func Bind(genPkg string, pkg *model.Package, sdkVersion string, isMainDalf bool, generateHexCodec bool, fieldHints ...model.FieldHints) (string, error) {
	hints := model.FieldHints{}
	if len(fieldHints) > 0 {
		hints = fieldHints[0]
	}

	// Collect all types used as choice arguments
	choiceArgTypes := make(map[string]bool)
	choiceArgChoices := make(map[string]string)
	choiceNameCounts := make(map[string]int)
	paramEncoderNames := make(map[string]bool)
	for _, tmpl := range pkg.Structs {
		if tmpl.IsTemplate {
			for _, choice := range tmpl.Choices {
				paramEncoderNames[choice.Name] = true
				choiceNameCounts[choice.Name]++
				if choice.ArgType != nil && choice.ArgType.GoType() != "" && choice.ArgType.GoType() != "UNIT" {
					argType := choice.ArgType.GoType()
					// Special case: SET type uses the choice name as the struct name
					// This matches the transformation done in the template for choice methods
					if argType == "SET" {
						argType = capitalize(choice.Name)
					}
					choiceArgTypes[argType] = true
					choiceArgChoices[argType] = choice.Name
				}
			}
		}
	}
	for name, enabled := range hints.ChoiceParamEncoderNames {
		if enabled {
			paramEncoderNames[name] = true
		}
	}
	encoderMethodNames := buildEncoderMethodNames(pkg.Structs, choiceArgTypes, choiceArgChoices, paramEncoderNames, choiceNameCounts)

	data := &tmplData{
		Package:            genPkg,
		PackageID:          pkg.PackageID,
		PackageName:        pkg.Name,
		SdkVersion:         sdkVersion,
		Structs:            pkg.Structs,
		IsMainDalf:         isMainDalf,
		GenerateHexCodec:   generateHexCodec,
		ChoiceArgTypes:     choiceArgTypes,
		ChoiceArgChoices:   choiceArgChoices,
		ParamEncoderNames:  paramEncoderNames,
		EncoderMethodNames: encoderMethodNames,
		ImportedPackages:   pkg.ImportedPackages,
	}
	buffer := new(bytes.Buffer)

	funcs := map[string]interface{}{
		"capitalise":        capitalize,
		"decapitalize":      decapitalize,
		"stringsHasPrefix":  strings.HasPrefix,
		"stringsTrimPrefix": strings.TrimPrefix,
		"stringsHasSuffix":  strings.HasSuffix,
		"stringsTrimSuffix": strings.TrimSuffix,
		"isGenMapType": func(t model.DamlType) bool {
			_, ok := t.(model.GenMap)
			return ok
		},
		"isTextMapType": func(t model.DamlType) bool {
			_, ok := t.(model.TextMap)
			return ok
		},
		"hasCallerField": func(s *model.TmplStruct) bool {
			for _, f := range s.Fields {
				if strings.EqualFold(f.Name, "caller") {
					return true
				}
			}
			return false
		},
		"isCallerField": func(fieldName string) bool {
			return strings.EqualFold(fieldName, "caller")
		},
		"damlName": func(s *model.TmplStruct) string {
			if s.DAMLName != "" {
				return s.DAMLName
			}
			return s.Name
		},
		// hasKey checks if a key exists in a map[string]byte (for VariantTagMapping)
		"hasKey": func(m map[string]byte, key string) bool {
			_, ok := m[key]
			return ok
		},
	}
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplSource))
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}
	// Pass the code through gofmt to clean it up
	code, err := format.Source(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, buffer)
	}
	return string(code), nil
}

func buildEncoderMethodNames(
	structs map[string]*model.TmplStruct,
	choiceArgTypes map[string]bool,
	choiceArgChoices map[string]string,
	paramEncoderNames map[string]bool,
	choiceNameCounts map[string]int,
) map[string]string {
	names := make(map[string]string)
	used := make(map[string]bool)
	structNames := sortedStructNames(structs)

	for _, structName := range structNames {
		s := structs[structName]
		if !isEncoderRecord(s) {
			continue
		}

		goName := capitalize(s.Name)
		if !choiceArgTypes[goName] {
			continue
		}

		choiceName := choiceArgChoices[goName]
		methodName := goName
		if choiceName != "" && choiceNameCounts[choiceName] == 1 {
			methodName = capitalize(choiceName)
		}
		names[goName] = reserveMethodName(methodName, goName, used)
	}

	for _, structName := range structNames {
		s := structs[structName]
		if !isEncoderRecord(s) {
			continue
		}

		goName := capitalize(s.Name)
		if _, ok := names[goName]; ok {
			continue
		}

		damlName := s.DAMLName
		if damlName == "" {
			damlName = s.Name
		}
		if !strings.HasSuffix(damlName, "Params") {
			continue
		}

		baseName := strings.TrimSuffix(damlName, "Params")
		if !paramEncoderNames[baseName] {
			continue
		}

		preferred := capitalize(baseName)
		fallback := preferred + "Params"
		names[goName] = reserveMethodName(preferred, fallback, used)
	}

	return names
}

func sortedStructNames(structs map[string]*model.TmplStruct) []string {
	names := make([]string, 0, len(structs))
	for name := range structs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func isEncoderRecord(s *model.TmplStruct) bool {
	return s != nil && !s.IsInterface && !s.IsTemplate && s.RawType == "Record"
}

func reserveMethodName(preferred string, fallback string, used map[string]bool) string {
	for _, name := range []string{preferred, fallback} {
		if name == "" || used[name] {
			continue
		}
		used[name] = true
		return name
	}

	base := fallback
	if base == "" {
		base = preferred
	}
	for i := 2; ; i++ {
		name := fmt.Sprintf("%s%d", base, i)
		if !used[name] {
			used[name] = true
			return name
		}
	}
}

func capitalize(input string) string {
	if len(input) == 0 {
		return input
	}

	hasSeparators := strings.ContainsAny(input, "_- ")

	if !hasSeparators && len(input) > 0 && input[0] >= 'A' && input[0] <= 'Z' {
		return input
	}

	result := toCamelCase(input)
	return strings.ToUpper(result[:1]) + result[1:]
}

func decapitalize(input string) string {
	if len(input) == 0 {
		return input
	}

	if isAllCaps(input) {
		return strings.ToLower(input)
	}

	if len(input) > 0 && input[0] >= 'a' && input[0] <= 'z' && !strings.ContainsAny(input, "_- ") {
		return input
	}

	result := toCamelCase(input)
	return strings.ToLower(result[:1]) + result[1:]
}

func toCamelCase(input string) string {
	if len(input) == 0 {
		return input
	}

	if !strings.ContainsAny(input, "_- ") {
		return input
	}

	words := strings.FieldsFunc(input, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	if len(words) == 0 {
		return input
	}

	var result strings.Builder
	for i, word := range words {
		if len(word) == 0 {
			continue
		}

		if isAllCaps(word) {
			if len(word) <= 3 {
				result.WriteString(word)
			} else {
				if i == 0 {
					result.WriteString(strings.ToLower(word))
				} else {
					result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
				}
			}
		} else {
			if i == 0 {
				result.WriteString(strings.ToLower(word[:1]) + word[1:])
			} else {
				result.WriteString(strings.ToUpper(word[:1]) + word[1:])
			}
		}
	}

	return result.String()
}

func isAllCaps(input string) bool {
	if len(input) == 0 {
		return false
	}
	for _, r := range input {
		if r >= 'a' && r <= 'z' {
			return false
		}
	}
	for _, r := range input {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}
