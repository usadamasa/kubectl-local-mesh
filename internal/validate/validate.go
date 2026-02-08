package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/usadamasa/kubectl-localmesh/schemas"
	"gopkg.in/yaml.v3"
)

// ValidationResult holds the results of schema validation.
type ValidationResult struct {
	Errors []string
}

// OK returns true if no validation errors were found.
func (r *ValidationResult) OK() bool {
	return len(r.Errors) == 0
}

// ValidateSchemaFile validates a YAML config file against the embedded JSON Schema.
func ValidateSchemaFile(path string) (*ValidationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Convert YAML-specific types to JSON-compatible types
	doc = convertYAMLToJSON(doc)

	return validateDocument(doc)
}

func validateDocument(doc any) (*ValidationResult, error) {
	compiler := jsonschema.NewCompiler()

	schemaDoc, err := jsonschema.UnmarshalJSON(strings.NewReader(schemas.ConfigSchema))
	if err != nil {
		return nil, fmt.Errorf("parsing embedded schema: %w", err)
	}

	if err := compiler.AddResource("config.schema.json", schemaDoc); err != nil {
		return nil, fmt.Errorf("loading schema: %w", err)
	}

	schema, err := compiler.Compile("config.schema.json")
	if err != nil {
		return nil, fmt.Errorf("compiling schema: %w", err)
	}

	result := &ValidationResult{}
	if err := schema.Validate(doc); err != nil {
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			collectErrors(ve, result)
		} else {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result, nil
}

func collectErrors(ve *jsonschema.ValidationError, result *ValidationResult) {
	if len(ve.Causes) == 0 {
		msg := ve.Error()
		result.Errors = append(result.Errors, msg)
		return
	}
	for _, cause := range ve.Causes {
		collectErrors(cause, result)
	}
}

// convertYAMLToJSON converts YAML-specific types to JSON-compatible types.
// yaml.v3 decodes integer values as int, but JSON Schema validation expects
// float64 for numeric values (matching encoding/json conventions).
func convertYAMLToJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = convertYAMLToJSON(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = convertYAMLToJSON(v)
		}
		return result
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return val
	}
}
