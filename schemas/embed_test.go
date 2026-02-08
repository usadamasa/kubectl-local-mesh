package schemas

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestConfigSchema_ValidJSON(t *testing.T) {
	// スキーマファイルがvalid JSONであることを確認
	var doc any
	if err := json.Unmarshal([]byte(ConfigSchema), &doc); err != nil {
		t.Fatalf("config.schema.json is not valid JSON: %v", err)
	}
}

func TestConfigSchema_Compilable(t *testing.T) {
	// スキーマがjsonschemaライブラリでコンパイルできることを確認
	compiler := jsonschema.NewCompiler()

	schemaDoc, err := jsonschema.UnmarshalJSON(strings.NewReader(ConfigSchema))
	if err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	if err := compiler.AddResource("config.schema.json", schemaDoc); err != nil {
		t.Fatalf("failed to add schema resource: %v", err)
	}

	_, err = compiler.Compile("config.schema.json")
	if err != nil {
		t.Fatalf("config.schema.json failed to compile: %v", err)
	}
}

func TestConfigSchema_HasRequiredStructure(t *testing.T) {
	// スキーマに必要な構造が含まれていることを確認
	var schema map[string]any
	if err := json.Unmarshal([]byte(ConfigSchema), &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// $schema フィールド
	schemaVersion, ok := schema["$schema"].(string)
	if !ok {
		t.Fatal("missing $schema field")
	}
	if schemaVersion != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("expected Draft 2020-12 schema, got %s", schemaVersion)
	}

	// type: object
	if schema["type"] != "object" {
		t.Errorf("expected root type 'object', got %v", schema["type"])
	}

	// required: services
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("missing required field")
	}
	found := false
	for _, r := range required {
		if r == "services" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'services' should be in required fields")
	}

	// $defs should contain Service, KubernetesService, TCPService, SSHBastion
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatal("missing $defs")
	}

	expectedDefs := []string{"Service", "KubernetesService", "TCPService", "SSHBastion"}
	for _, name := range expectedDefs {
		if _, ok := defs[name]; !ok {
			t.Errorf("missing $defs/%s", name)
		}
	}
}

func TestConfigSchema_ServiceOneOf(t *testing.T) {
	// Service定義がoneOfでKubernetesServiceとTCPServiceを持つことを確認
	var schema map[string]any
	if err := json.Unmarshal([]byte(ConfigSchema), &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	defs := schema["$defs"].(map[string]any)
	service := defs["Service"].(map[string]any)

	oneOf, ok := service["oneOf"].([]any)
	if !ok {
		t.Fatal("Service should have oneOf")
	}
	if len(oneOf) != 2 {
		t.Errorf("Service oneOf should have 2 items, got %d", len(oneOf))
	}
}

func TestConfigSchema_KubernetesServiceKindEnum(t *testing.T) {
	// KubernetesServiceのkindがenum ["kubernetes"]であることを確認
	var schema map[string]any
	if err := json.Unmarshal([]byte(ConfigSchema), &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	defs := schema["$defs"].(map[string]any)
	k8s := defs["KubernetesService"].(map[string]any)
	props := k8s["properties"].(map[string]any)
	kind := props["kind"].(map[string]any)

	enum, ok := kind["enum"].([]any)
	if !ok || len(enum) != 1 {
		t.Fatalf("KubernetesService kind should have enum with 1 value, got %v", kind["enum"])
	}
	if enum[0] != "kubernetes" {
		t.Errorf("KubernetesService kind enum should be ['kubernetes'], got %v", enum)
	}
}

func TestConfigSchema_TCPServiceKindEnum(t *testing.T) {
	// TCPServiceのkindがenum ["tcp"]であることを確認
	var schema map[string]any
	if err := json.Unmarshal([]byte(ConfigSchema), &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	defs := schema["$defs"].(map[string]any)
	tcp := defs["TCPService"].(map[string]any)
	props := tcp["properties"].(map[string]any)
	kind := props["kind"].(map[string]any)

	enum, ok := kind["enum"].([]any)
	if !ok || len(enum) != 1 {
		t.Fatalf("TCPService kind should have enum with 1 value, got %v", kind["enum"])
	}
	if enum[0] != "tcp" {
		t.Errorf("TCPService kind enum should be ['tcp'], got %v", enum)
	}
}

func TestConfigSchema_AdditionalPropertiesFalse(t *testing.T) {
	// 全objectにadditionalProperties: falseが設定されていることを確認
	var schema map[string]any
	if err := json.Unmarshal([]byte(ConfigSchema), &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	// Root object
	if schema["additionalProperties"] != false {
		t.Error("root object should have additionalProperties: false")
	}

	defs := schema["$defs"].(map[string]any)

	// KubernetesService
	k8s := defs["KubernetesService"].(map[string]any)
	if k8s["additionalProperties"] != false {
		t.Error("KubernetesService should have additionalProperties: false")
	}

	// TCPService
	tcp := defs["TCPService"].(map[string]any)
	if tcp["additionalProperties"] != false {
		t.Error("TCPService should have additionalProperties: false")
	}

	// SSHBastion
	bastion := defs["SSHBastion"].(map[string]any)
	if bastion["additionalProperties"] != false {
		t.Error("SSHBastion should have additionalProperties: false")
	}
}

func TestConfigSchema_ProtocolEnum(t *testing.T) {
	// protocolフィールドのenum値を確認
	var schema map[string]any
	if err := json.Unmarshal([]byte(ConfigSchema), &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	defs := schema["$defs"].(map[string]any)
	k8s := defs["KubernetesService"].(map[string]any)
	props := k8s["properties"].(map[string]any)
	protocol := props["protocol"].(map[string]any)

	enum, ok := protocol["enum"].([]any)
	if !ok {
		t.Fatal("protocol should have enum")
	}

	expected := map[string]bool{"http": true, "http2": true, "grpc": true}
	for _, v := range enum {
		s, ok := v.(string)
		if !ok {
			t.Errorf("enum value should be string, got %T", v)
			continue
		}
		if !expected[s] {
			t.Errorf("unexpected protocol enum value: %s", s)
		}
		delete(expected, s)
	}
	for missing := range expected {
		t.Errorf("missing protocol enum value: %s", missing)
	}
}
