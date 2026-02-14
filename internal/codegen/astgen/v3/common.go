package v3

import (
	"errors"
	"fmt"
	"strings"

	damlcommon "github.com/digital-asset/dazl-client/v8/go/api/com/digitalasset/daml/lf/archive"
	daml "github.com/digital-asset/dazl-client/v8/go/api/com/digitalasset/daml/lf/archive/daml_lf_2"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/go-daml/internal/codegen/model"
	"google.golang.org/protobuf/proto"
)

const (
	RawTypeTemplate   = "Template"
	RawTypeOptional   = "OPTIONAL"
	RawTypeInterface  = "Interface"
	RawTypeRecord     = "Record"
	RawTypeVariant    = "Variant"
	RawTypeEnum       = "Enum"
	RawTypeContractID = "CONTRACT_ID"
	RawTypeList       = "LIST"
)

type codeGenAst struct {
	payload []byte
}

func NewCodegenAst(payload []byte) *codeGenAst {
	return &codeGenAst{payload: payload}
}

func (c *codeGenAst) isEnumType(typeName string, pkg *daml.Package) bool {
	for _, module := range pkg.Modules {
		for _, dataType := range module.GetDataTypes() {
			if !dataType.Serializable {
				continue
			}

			name := c.getName(pkg, dataType.GetNameInternedDname())
			if name == typeName {
				if _, isEnum := dataType.DataCons.(*daml.DefDataType_Enum); isEnum {
					return true
				}
			}
		}
	}
	return false
}

func (c *codeGenAst) GetInterfaces() (map[string]*model.TmplStruct, error) {
	interfaceMap := make(map[string]*model.TmplStruct)

	var archive damlcommon.Archive
	err := proto.Unmarshal(c.payload, &archive)
	if err != nil {
		return nil, err
	}

	var payloadMapped damlcommon.ArchivePayload
	err = proto.Unmarshal(archive.Payload, &payloadMapped)
	if err != nil {
		return nil, err
	}

	damlLfBytes := payloadMapped.GetDamlLf_2()
	if damlLfBytes == nil {
		return nil, errors.New("unsupported daml version")
	}

	var damlLf daml.Package
	err = proto.Unmarshal(damlLfBytes, &damlLf)
	if err != nil {
		return nil, err
	}

	for _, module := range damlLf.Modules {
		if len(damlLf.InternedStrings) == 0 {
			continue
		}

		moduleName := c.getDottedName(&damlLf, module.GetNameInternedDname())

		interfaces, err := c.getInterfaces(&damlLf, module, moduleName)
		if err != nil {
			return nil, err
		}
		for key, val := range interfaces {
			interfaceMap[key] = val
		}
	}

	return interfaceMap, nil
}

func (c *codeGenAst) GetTemplateStructs(ifcByModule map[string]model.InterfaceMap) (map[string]*model.TmplStruct, error) {
	structs := make(map[string]*model.TmplStruct)

	var archive damlcommon.Archive
	err := proto.Unmarshal(c.payload, &archive)
	if err != nil {
		return nil, err
	}

	var payloadMapped damlcommon.ArchivePayload
	err = proto.Unmarshal(archive.Payload, &payloadMapped)
	if err != nil {
		return nil, err
	}

	damlLfBytes := payloadMapped.GetDamlLf_2()
	if damlLfBytes == nil {
		return nil, errors.New("unsupported daml version")
	}

	var damlLf daml.Package
	err = proto.Unmarshal(damlLfBytes, &damlLf)
	if err != nil {
		return nil, err
	}

	for _, module := range damlLf.Modules {
		if len(damlLf.InternedStrings) == 0 {
			continue
		}

		moduleName := c.getDottedName(&damlLf, module.GetNameInternedDname())
		log.Info().Msgf("processing module %s", moduleName)

		dataTypes, err := c.getDataTypes(&damlLf, module, moduleName)
		if err != nil {
			return nil, err
		}
		for key, val := range dataTypes {
			structs[key] = val
		}

		templates, err := c.getTemplates(&damlLf, module, moduleName, ifcByModule)
		if err != nil {
			return nil, err
		}
		for key, val := range templates {
			structs[key] = val
		}

	}

	return structs, nil
}

func (c *codeGenAst) getName(pkg *daml.Package, id int32) string {
	idx := pkg.InternedDottedNames[id].SegmentsInternedStr
	return pkg.InternedStrings[idx[len(idx)-1]]
}

func (c *codeGenAst) getTemplates(
	pkg *daml.Package, module *daml.Module, moduleName string,
	interfaces map[string]model.InterfaceMap,
) (map[string]*model.TmplStruct, error) {
	structs := make(map[string]*model.TmplStruct, 0)

	for _, template := range module.Templates {
		templateName := c.getName(pkg, template.TyconInternedDname)
		log.Debug().Msgf("processing template: %s", templateName)

		var templateDataType *daml.DefDataType
		for _, dataType := range module.DataTypes {
			dtName := c.getName(pkg, dataType.GetNameInternedDname())
			if dtName == templateName {
				templateDataType = dataType
				break
			}
		}

		if templateDataType == nil {
			log.Debug().Msgf("could not find data type for template: %s", templateName)
			continue
		}

		tmplStruct := model.TmplStruct{
			Name:       templateName,
			ModuleName: moduleName,
			RawType:    RawTypeTemplate,
			IsTemplate: true,
			Choices:    make([]*model.TmplChoice, 0),
		}

		switch v := templateDataType.DataCons.(type) {
		case *daml.DefDataType_Record:
			for _, field := range v.Record.Fields {
				fieldExtracted, typeExtracted, err := c.extractField(pkg, field)
				if err != nil {
					return nil, err
				}
				isOptional := typeExtracted == RawTypeOptional || strings.HasPrefix(typeExtracted, "*")
				tmplStruct.Fields = append(tmplStruct.Fields, &model.TmplField{
					Name:       fieldExtracted,
					Type:       typeExtracted,
					RawType:    field.String(),
					IsOptional: isOptional,
					IsEnum:     c.isEnumType(typeExtracted, pkg),
				})
			}
		default:
			log.Debug().Msgf("template %s has non-record data type: %T", templateName, v)
		}

		choices := c.getChoices(pkg, template.Choices)
		tmplStruct.Choices = append(tmplStruct.Choices, choices...)

		if template.Key != nil {
			keyType := template.Key.GetType().String()
			normalizedKeyType := model.NormalizeDAMLType(keyType)
			log.Debug().Msgf("template %s has key of type: %s (normalized: %s)", templateName, keyType, normalizedKeyType)
			keyFieldNames := c.parseKeyExpression(pkg, template.Key)

			if len(keyFieldNames) > 0 {
				// For now, we support single-field keys
				// TODO: Support composite keys with multiple fields
				keyFieldName := keyFieldNames[0]
				var keyField *model.TmplField
				for _, field := range tmplStruct.Fields {
					if field.Name == keyFieldName {
						keyField = &model.TmplField{
							Name:    field.Name,
							Type:    field.Type,
							RawType: keyType,
						}
						break
					}
				}

				if keyField != nil {
					tmplStruct.Key = keyField
					log.Debug().Msgf("template %s key field: %s", templateName, keyFieldName)
				}
			}
		}

		if len(template.Implements) > 0 {
			for _, impl := range template.Implements {
				if impl.Interface != nil {
					interfaceName := "I" + c.getName(pkg, impl.Interface.GetNameInternedDname())
					tmplStruct.Implements = append(tmplStruct.Implements, interfaceName)
					ifcModuleName := c.getDottedName(pkg, impl.Interface.Module.ModuleNameInternedDname)
					log.Debug().Msgf("template %s -implements interface: %s location %s", templateName, interfaceName, ifcModuleName)

					if interfaceStruct, exists := interfaces[ifcModuleName][interfaceName]; exists {
						log.Debug().Msgf("found interface %s in map with %d choices", interfaceName, len(interfaceStruct.Choices))
						for _, ifaceChoice := range interfaceStruct.Choices {
							found := false
							for _, tmplChoice := range tmplStruct.Choices {
								if tmplChoice.Name == ifaceChoice.Name {
									found = true
									break
								}
							}
							if !found {
								log.Debug().Msgf("adding interface choice %s to template %s", ifaceChoice.Name, templateName)
								tmplStruct.Choices = append(tmplStruct.Choices, &model.TmplChoice{
									Name:              ifaceChoice.Name,
									ArgType:           ifaceChoice.ArgType,
									ReturnType:        ifaceChoice.ReturnType,
									InterfaceName:     interfaceName,
									InterfaceDAMLName: interfaceStruct.DAMLName,
								})
							}
						}
					}
				}
			}
		}

		structs[templateName] = &tmplStruct
	}

	return structs, nil
}

func (c *codeGenAst) getChoices(pkg *daml.Package, choices []*daml.TemplateChoice) []*model.TmplChoice {
	res := make([]*model.TmplChoice, 0)
	for _, choice := range choices {
		choiceName := pkg.InternedStrings[choice.NameInternedStr]
		choiceStruct := &model.TmplChoice{
			Name: choiceName,
		}

		// Extract argument type if present
		if argBinder := choice.GetArgBinder(); argBinder != nil && argBinder.Type != nil {
			argType := c.extractType(pkg, argBinder.Type)
			// Only set ArgType if it's not a UNIT type
			if argType != "UNIT" && argType != "" {
				choiceStruct.ArgType = argType
			}
		}

		if retType := choice.GetRetType(); retType != nil {
			choiceStruct.ReturnType = c.extractType(pkg, retType)
		}

		res = append(res, choiceStruct)
	}

	return res
}

func (c *codeGenAst) getInterfaces(pkg *daml.Package, module *daml.Module, moduleName string) (map[string]*model.TmplStruct, error) {
	structs := make(map[string]*model.TmplStruct, 0)

	for _, iface := range module.Interfaces {
		originalName := c.getName(pkg, iface.TyconInternedDname)
		interfaceName := "I" + originalName
		location := c.getName(pkg, iface.Location.Module.GetModuleNameInternedDname())
		log.Debug().Msgf("processing interface: %s, original name %s location %s", interfaceName, originalName, location)

		tmplStruct := model.TmplStruct{
			Name:        interfaceName,
			DAMLName:    originalName,
			ModuleName:  moduleName,
			RawType:     RawTypeInterface,
			IsInterface: true, // TODO dont need as we have RawTypeInterface
			Choices:     make([]*model.TmplChoice, 0),
			Location:    location,
		}
		choices := c.getChoices(pkg, iface.Choices)
		tmplStruct.Choices = append(tmplStruct.Choices, choices...)

		// TODO: Process interface view if needed
		// iface.View contains the view type information

		structs[interfaceName] = &tmplStruct
	}

	return structs, nil
}

func (c *codeGenAst) getDataTypes(pkg *daml.Package, module *daml.Module, moduleName string) (map[string]*model.TmplStruct, error) {
	structs := make(map[string]*model.TmplStruct, 0)
	for _, dataType := range module.GetDataTypes() {
		if !dataType.Serializable {
			continue
		}

		name := c.getName(pkg, dataType.GetNameInternedDname())
		tmplStruct := model.TmplStruct{
			Name:       name,
			ModuleName: moduleName,
		}

		switch v := dataType.DataCons.(type) {
		case *daml.DefDataType_Record:
			tmplStruct.RawType = RawTypeRecord
			for _, field := range v.Record.Fields {
				fieldExtracted, typeExtracted, err := c.extractField(pkg, field)
				if err != nil {
					return nil, err
				}
				isOptional := typeExtracted == RawTypeOptional || strings.HasPrefix(typeExtracted, "*")
				tmplStruct.Fields = append(tmplStruct.Fields, &model.TmplField{
					Name:       fieldExtracted,
					Type:       typeExtracted,
					RawType:    field.String(),
					IsOptional: isOptional,
				})
			}
		case *daml.DefDataType_Variant:
			tmplStruct.RawType = RawTypeVariant
			for _, field := range v.Variant.Fields {
				fieldExtracted, typeExtracted, err := c.extractField(pkg, field)
				if err != nil {
					return nil, err
				}
				tmplStruct.Fields = append(tmplStruct.Fields, &model.TmplField{
					Name:       fieldExtracted,
					Type:       typeExtracted,
					RawType:    field.String(),
					IsOptional: true,
				})
			}
		case *daml.DefDataType_Enum:
			tmplStruct.RawType = RawTypeEnum
			for _, constructorIdx := range v.Enum.ConstructorsInternedStr {
				if int(constructorIdx) < len(pkg.InternedStrings) {
					constructorName := pkg.InternedStrings[constructorIdx]
					tmplStruct.Fields = append(tmplStruct.Fields, &model.TmplField{
						Name: constructorName,
						Type: "enum",
					})
				}
			}
		case *daml.DefDataType_Interface:
			tmplStruct.RawType = RawTypeInterface
			log.Warn().Msgf("interface not supported %s", v.Interface.String())
		default:
			log.Warn().Msgf("unknown data cons type: %T", v)
		}
		structs[name] = &tmplStruct
	}

	return structs, nil
}

func (c *codeGenAst) parseKeyExpression(pkg *daml.Package, key *daml.DefTemplate_DefKey) []string {
	var fieldNames []string
	if key == nil || key.KeyExpr == nil {
		return fieldNames
	}
	fieldNames = c.parseExpressionForFields(pkg, key.KeyExpr)

	if len(fieldNames) == 0 {
		log.Warn().Msg("could not extract fields from key expression")
	}

	return fieldNames
}

func (c *codeGenAst) parseExpressionForFields(pkg *daml.Package, expr *daml.Expr) []string {
	var fieldNames []string

	if expr == nil {
		return fieldNames
	}

	switch e := expr.Sum.(type) {
	case *daml.Expr_RecProj_:
		if e.RecProj != nil {
			if e.RecProj.FieldInternedStr != 0 {
				fieldName := pkg.InternedStrings[e.RecProj.FieldInternedStr]
				fieldNames = append(fieldNames, fieldName)
			}
			// Also check if the record being projected has more fields
			if e.RecProj.Record != nil {
				subFields := c.parseExpressionForFields(pkg, e.RecProj.Record)
				fieldNames = append(fieldNames, subFields...)
			}
		}
	case *daml.Expr_RecCon_:
		if e.RecCon != nil {
			for _, field := range e.RecCon.Fields {
				if field.FieldInternedStr != 0 {
					fieldName := pkg.InternedStrings[field.FieldInternedStr]
					fieldNames = append(fieldNames, fieldName)
				}
			}
		}
	case *daml.Expr_VarInternedStr:
		if e.VarInternedStr != 0 {
			varName := pkg.InternedStrings[e.VarInternedStr]
			// In template keys, the template parameter is often referenced
			// We'll include variable names as they might represent fields
			fieldNames = append(fieldNames, varName)
		}
	case *daml.Expr_Builtin:
		// Builtin function - might have arguments with field references
		// In DAML LF 2.1, builtins are handled differently
		// For now, we don't extract fields from builtins
	default:
		log.Debug().Msgf("unhandled expression type in key parsing: %T", e)
	}

	return fieldNames
}

func (c *codeGenAst) extractTapp(pkg *daml.Package, tapp *daml.Type_TApp) string {
	if tapp == nil {
		return "unknown_tapp"
	}

	lhs := model.NormalizeDAMLType(c.extractType(pkg, tapp.GetLhs()))

	switch lhs {
	case "LIST":
		elem := model.NormalizeDAMLType(c.extractType(pkg, tapp.GetRhs()))
		return "[]" + elem

	case "OPTIONAL":
		elem := model.NormalizeDAMLType(c.extractType(pkg, tapp.GetRhs()))
		return "*" + elem

	case "CONTRACT_ID":
		// ContractId X  -> CONTRACT_ID (don’t collapse to string)
		return "CONTRACT_ID"
	}

	// some other type application; keep lhs
	return lhs
}
func (c *codeGenAst) extractType(pkg *daml.Package, typ *daml.Type) string {
	if typ == nil {
		return ""
	}

	switch v := typ.Sum.(type) {

	case *daml.Type_InternedType:
		prim := pkg.InternedTypes[v.InternedType]
		if prim == nil {
			return "unknown_interned_type"
		}
		// recurse into the interned definition (may be TApp/Builtin/Con/etc)
		return model.NormalizeDAMLType(c.extractType(pkg, prim))

	case *daml.Type_Tapp: // ✅ MUST be Type_Tapp? nope -> Type_Tapp is wrong; you want Type_Tapp? actually the oneof wrapper is Type_Tapp, but the inner is Type_TApp
		// IMPORTANT: depending on generated code, wrapper is Type_Tapp but inner is Type_TApp.
		// In your earlier error, v.Tapp is *Type_TApp, so this compiles.
		return model.NormalizeDAMLType(c.extractTapp(pkg, v.Tapp))

	case *daml.Type_Builtin_:
		return model.NormalizeDAMLType(c.handleBuiltinType(pkg, v.Builtin))

	case *daml.Type_Con_:
		if v.Con.Tycon != nil {
			return model.NormalizeDAMLType(c.getName(pkg, v.Con.Tycon.GetNameInternedDname()))
		}
		return "con_without_tycon"

	case *daml.Type_Var_:
		if int(v.Var.GetVarInternedStr()) < len(pkg.InternedStrings) {
			return model.NormalizeDAMLType(pkg.InternedStrings[v.Var.GetVarInternedStr()])
		}
		return "unknown_var"

	case *daml.Type_Syn_:
		if v.Syn.Tysyn != nil {
			return model.NormalizeDAMLType(fmt.Sprintf("syn_%s", c.getName(pkg, v.Syn.Tysyn.GetNameInternedDname())))
		}
		return "syn_without_name"

	default:
		return model.NormalizeDAMLType(fmt.Sprintf("unknown_type_%T", typ.Sum))
	}
}

func (c *codeGenAst) handleBuiltinType(pkg *daml.Package, b *daml.Type_Builtin) string {
	switch b.Builtin {

	case daml.BuiltinType_LIST:
		// LIST a
		if len(b.Args) > 0 {
			elem := model.NormalizeDAMLType(c.extractType(pkg, b.Args[0]))
			return "[]" + elem
		}
		return RawTypeList // fallback

	case daml.BuiltinType_OPTIONAL:
		// Optional a
		if len(b.Args) > 0 {
			elem := model.NormalizeDAMLType(c.extractType(pkg, b.Args[0]))
			return "*" + elem
		}
		return RawTypeOptional

	case daml.BuiltinType_CONTRACT_ID:
		// ContractId a  -> CONTRACT_ID (do not collapse to string)
		return RawTypeContractID

	case daml.BuiltinType_GENMAP:
		return "GENMAP"

	case daml.BuiltinType_TEXTMAP:
		return "TEXTMAP"

	default:
		return b.Builtin.String()
	}
}

func (c *codeGenAst) handleConType(pkg *daml.Package, conType *daml.Type_Con) string {
	if conType == nil || conType.Tycon == nil {
		return "con_without_tycon"
	}

	tyconName := c.getName(pkg, conType.Tycon.GetNameInternedDname())

	switch tyconName {
	case "Optional":
		if len(conType.Args) > 0 {
			elementType := c.extractType(pkg, conType.Args[0])
			normalizedElementType := model.NormalizeDAMLType(elementType)
			return "*" + normalizedElementType
		}
		return RawTypeOptional
	case "List":
		if len(conType.Args) > 0 {
			elementType := c.extractType(pkg, conType.Args[0])
			normalizedElementType := model.NormalizeDAMLType(elementType)
			return "[]" + normalizedElementType
		}
		return RawTypeList
	case "Tuple2":
		if len(conType.Args) >= 2 {
			arg1Type := c.extractType(pkg, conType.Args[0])
			arg2Type := c.extractType(pkg, conType.Args[1])
			return "TUPLE2[" + model.NormalizeDAMLType(arg1Type) + "," + model.NormalizeDAMLType(arg2Type) + "]"
		}
		return "TUPLE2"
	case "Tuple3":
		if len(conType.Args) >= 3 {
			arg1Type := c.extractType(pkg, conType.Args[0])
			arg2Type := c.extractType(pkg, conType.Args[1])
			arg3Type := c.extractType(pkg, conType.Args[2])
			return "TUPLE3[" + model.NormalizeDAMLType(arg1Type) + "," + model.NormalizeDAMLType(arg2Type) + "," + model.NormalizeDAMLType(arg3Type) + "]"
		}
		return "TUPLE3"
	default:
		return tyconName
	}
}

func (c *codeGenAst) extractField(pkg *daml.Package, field *daml.FieldWithType) (string, string, error) {
	if field == nil {
		return "", "", fmt.Errorf("field is nil")
	}

	idx := field.GetFieldInternedStr()
	if int(idx) >= len(pkg.InternedStrings) {
		return "", "", fmt.Errorf("invalid interned string index for field name: %d", idx)
	}
	fieldName := pkg.InternedStrings[idx]

	if field.Type == nil {
		return fieldName, "", fmt.Errorf("field type is nil")
	}

	ty := c.extractType(pkg, field.Type) // ✅ funnels everything through tapp/builtin/etc
	return fieldName, model.NormalizeDAMLType(ty), nil
}

func (c *codeGenAst) getDottedName(pkg *daml.Package, dottedNameID int32) string {
	if int(dottedNameID) >= len(pkg.InternedDottedNames) {
		return ""
	}
	segments := pkg.InternedDottedNames[dottedNameID].SegmentsInternedStr
	if len(segments) == 0 {
		return ""
	}

	parts := make([]string, 0, len(segments))
	for _, segIdx := range segments {
		if int(segIdx) < len(pkg.InternedStrings) {
			parts = append(parts, pkg.InternedStrings[segIdx])
		}
	}
	return strings.Join(parts, ".")
}
