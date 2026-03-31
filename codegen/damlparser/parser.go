// Package damlparser provides a lightweight parser for Daml *Types.daml source files.
// It extracts record and variant type definitions for codec generation.
package damlparser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ParsedModule represents a parsed Daml module with its type definitions.
type ParsedModule struct {
	ModuleName string
	Imports    []string
	Records    []ParsedRecord
	Variants   []ParsedVariant
}

// ParsedRecord represents a Daml record type (data Name = Name with ...).
type ParsedRecord struct {
	Name   string
	Fields []ParsedField
}

// ParsedField represents a field in a record type.
type ParsedField struct {
	Name     string
	TypeExpr string // e.g., "Numeric 0", "Bool", "[RawInstanceAddress]"
}

// ParsedVariant represents a Daml variant/sum type (data Name = Ctor1 | Ctor2 Payload).
type ParsedVariant struct {
	Name         string
	Constructors []ParsedConstructor
}

// ParsedConstructor represents a constructor in a variant type.
type ParsedConstructor struct {
	Name        string
	PayloadType string // empty for unit constructors like "Indefinite"
}

// Regex patterns for parsing Daml source
var (
	// Module declaration: module MyModule.Types where
	// Also handles: module MyModule.Types (with export list on next lines)
	modulePattern          = regexp.MustCompile(`^module\s+(\S+)\s+where`)
	modulePatternMultiLine = regexp.MustCompile(`^module\s+(\S+)\s*$`)

	// Import patterns
	importPattern          = regexp.MustCompile(`^import\s+(\S+)`)
	importQualifiedPattern = regexp.MustCompile(`^import\s+(\S+)\s+qualified`)

	// Record start: data Name = Name
	// May have "with" on same line or next line
	recordStartPattern = regexp.MustCompile(`^data\s+(\w+)\s*=\s*(\w+)\s*(with)?`)

	// Variant start: data Name (no "= Name" pattern, has | separators)
	// First look for pattern like: data TransferTimeout
	variantStartPattern = regexp.MustCompile(`^data\s+(\w+)\s*$`)

	// Field pattern: fieldName : TypeExpr -- optional comment
	fieldPattern = regexp.MustCompile(`^\s+(\w+)\s*:\s*([^-]+)`)

	// Deriving pattern: deriving (Eq, Show)
	derivingPattern = regexp.MustCompile(`^\s*deriving\s+\(`)
)

// Parse reads a Daml source file and extracts type definitions.
func Parse(r io.Reader) (*ParsedModule, error) {
	scanner := bufio.NewScanner(r)
	module := &ParsedModule{}

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning input: %w", err)
	}

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			i++
			continue
		}

		// Module declaration
		if match := modulePattern.FindStringSubmatch(trimmed); match != nil {
			module.ModuleName = match[1]
			i++
			continue
		}

		// Multi-line module declaration with export list
		if match := modulePatternMultiLine.FindStringSubmatch(trimmed); match != nil {
			module.ModuleName = match[1]
			// Skip until we find "where"
			i++
			for i < len(lines) {
				if strings.Contains(lines[i], "where") {
					i++
					break
				}
				i++
			}
			continue
		}

		// Import
		if strings.HasPrefix(trimmed, "import") {
			if match := importQualifiedPattern.FindStringSubmatch(trimmed); match != nil {
				module.Imports = append(module.Imports, match[1])
			} else if match := importPattern.FindStringSubmatch(trimmed); match != nil {
				module.Imports = append(module.Imports, match[1])
			}
			i++
			continue
		}

		// Data type definition
		if strings.HasPrefix(trimmed, "data ") {
			record, variant, consumed, err := parseDataType(lines, i)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", i+1, err)
			}
			if record != nil {
				module.Records = append(module.Records, *record)
			}
			if variant != nil {
				module.Variants = append(module.Variants, *variant)
			}
			i += consumed
			continue
		}

		i++
	}

	return module, nil
}

// parseDataType parses a data type definition starting at lines[start].
// Returns the parsed record or variant, and the number of lines consumed.
func parseDataType(lines []string, start int) (*ParsedRecord, *ParsedVariant, int, error) {
	if start >= len(lines) {
		return nil, nil, 0, fmt.Errorf("unexpected end of input")
	}

	line := strings.TrimSpace(lines[start])

	// Check for record pattern: data Name = Name [with]
	if match := recordStartPattern.FindStringSubmatch(line); match != nil {
		typeName := match[1]
		ctorName := match[2]
		hasWithOnSameLine := match[3] == "with"

		// Verify it's a record (type name == constructor name)
		if typeName == ctorName {
			return parseRecord(lines, start, typeName, hasWithOnSameLine)
		}
		// Otherwise it might be a variant with a single-line definition
		return parseVariantSingleLine(lines, start, typeName)
	}

	// Check for multi-line variant: data Name\n  = Ctor1 | Ctor2
	if match := variantStartPattern.FindStringSubmatch(line); match != nil {
		typeName := match[1]
		return parseVariantMultiLine(lines, start, typeName)
	}

	// Check for inline variant: data Name = Ctor1 | Ctor2 Payload
	if strings.HasPrefix(line, "data ") && strings.Contains(line, "|") {
		return parseVariantInline(lines, start)
	}

	return nil, nil, 1, nil // Unknown pattern, skip
}

// parseRecord parses a record type definition.
func parseRecord(lines []string, start int, name string, hasWithOnSameLine bool) (*ParsedRecord, *ParsedVariant, int, error) {
	record := &ParsedRecord{Name: name}
	i := start + 1

	// If "with" wasn't on the first line, look for it
	if !hasWithOnSameLine && i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "with" {
			i++
		}
	}

	// Parse fields until we hit deriving or a non-field line
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comment-only lines
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			i++
			continue
		}

		// Check for deriving - end of record
		if derivingPattern.MatchString(trimmed) {
			i++
			break
		}

		// Check for field
		if match := fieldPattern.FindStringSubmatch(line); match != nil {
			fieldName := match[1]
			typeExpr := strings.TrimSpace(match[2])
			record.Fields = append(record.Fields, ParsedField{
				Name:     fieldName,
				TypeExpr: typeExpr,
			})
			i++
			continue
		}

		// Non-field line (could be next data definition or other)
		break
	}

	return record, nil, i - start, nil
}

// parseVariantMultiLine parses a variant with constructors on subsequent lines.
// Example:
//
//	data TransferTimeout
//	    = Indefinite
//	    | RelativeHours Int
//	    deriving (Eq, Show)
func parseVariantMultiLine(lines []string, start int, name string) (*ParsedRecord, *ParsedVariant, int, error) {
	variant := &ParsedVariant{Name: name}
	i := start + 1

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and pure comment lines
		if trimmed == "" || (strings.HasPrefix(trimmed, "--") && !strings.Contains(trimmed, "=") && !strings.Contains(trimmed, "|")) {
			i++
			continue
		}

		// Check for deriving - end of variant
		if derivingPattern.MatchString(trimmed) {
			i++
			break
		}

		// Parse constructor line (starts with = or |)
		if strings.HasPrefix(trimmed, "=") || strings.HasPrefix(trimmed, "|") {
			ctors := parseConstructorLine(trimmed)
			variant.Constructors = append(variant.Constructors, ctors...)
			i++
			continue
		}

		// Non-constructor line
		break
	}

	return nil, variant, i - start, nil
}

// parseVariantSingleLine parses a variant where everything is on one line after the =.
// Example: data TransferTimeout = Indefinite | RelativeHours Int deriving (Eq, Show)
func parseVariantSingleLine(lines []string, start int, name string) (*ParsedRecord, *ParsedVariant, int, error) {
	line := strings.TrimSpace(lines[start])
	variant := &ParsedVariant{Name: name}

	// Find the = and parse everything after it
	eqIdx := strings.Index(line, "=")
	if eqIdx == -1 {
		return nil, nil, 1, nil
	}

	rest := strings.TrimSpace(line[eqIdx+1:])

	// Remove deriving clause if present on same line
	if idx := strings.Index(rest, "deriving"); idx != -1 {
		rest = strings.TrimSpace(rest[:idx])
	}

	// Parse constructors separated by |
	ctors := parseConstructorLine("= " + rest)
	variant.Constructors = ctors

	// Check for deriving on same or next line
	consumed := 1
	if !strings.Contains(line, "deriving") && start+1 < len(lines) {
		if derivingPattern.MatchString(strings.TrimSpace(lines[start+1])) {
			consumed = 2
		}
	}

	return nil, variant, consumed, nil
}

// parseVariantInline parses a variant where constructors are on the same line.
func parseVariantInline(lines []string, start int) (*ParsedRecord, *ParsedVariant, int, error) {
	line := strings.TrimSpace(lines[start])

	// Extract type name: data TypeName = ...
	dataIdx := strings.Index(line, "data ")
	eqIdx := strings.Index(line, "=")
	if dataIdx == -1 || eqIdx == -1 {
		return nil, nil, 1, nil
	}

	name := strings.TrimSpace(line[dataIdx+5 : eqIdx])
	return parseVariantSingleLine(lines, start, name)
}

// parseConstructorLine parses a line containing variant constructors.
// Input: "= Indefinite | RelativeHours Int" or "| RelativeHours Int"
// Returns slice of ParsedConstructor.
func parseConstructorLine(line string) []ParsedConstructor {
	var ctors []ParsedConstructor

	// Remove leading = or |
	line = strings.TrimPrefix(strings.TrimSpace(line), "=")
	line = strings.TrimSpace(line)

	// Split by |
	parts := strings.Split(line, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Remove trailing comment
		if idx := strings.Index(part, "--"); idx != -1 {
			part = strings.TrimSpace(part[:idx])
		}

		// Split by space: CtorName [PayloadType]
		words := strings.Fields(part)
		if len(words) == 0 {
			continue
		}

		ctor := ParsedConstructor{Name: words[0]}
		if len(words) > 1 {
			ctor.PayloadType = strings.Join(words[1:], " ")
		}
		ctors = append(ctors, ctor)
	}

	return ctors
}

// ParseString parses Daml source from a string.
func ParseString(source string) (*ParsedModule, error) {
	return Parse(strings.NewReader(source))
}
