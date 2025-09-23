package codegen

import "time"

type tmplStruct struct {
	Name    string
	Fields  []*tmplField
	RawType string
}

type tmplField struct {
	Type       string
	Name       string
	RawType    string
	IsOptional bool
}

type Package struct {
	Name      string
	Version   string
	PackageID string
	Structs   map[string]*tmplStruct
	Metadata  *Metadata
}

type Metadata struct {
	Name         string
	Version      string
	Dependencies []string
	LangVersion  string
	CreatedBy    string
	SdkVersion   string
	CreatedAt    *time.Time
}
