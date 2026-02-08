package dump

import (
	"context"
	"testing"

	"github.com/usadamasa/kubectl-localmesh/internal/config"
)

func TestDumpVisitor_Creation(t *testing.T) {
	// DumpVisitorの生成テスト
	ctx := context.Background()

	visitor := NewDumpVisitor(ctx, "", nil)

	if visitor == nil {
		t.Fatal("expected visitor to be created")
	}

	if len(visitor.GetServiceConfigs()) != 0 {
		t.Errorf("expected 0 configs, got %d", len(visitor.GetServiceConfigs()))
	}
}

func TestDumpVisitor_VisitTCP(t *testing.T) {
	// DumpVisitorのVisitTCPテスト（モック不要）
	ctx := context.Background()

	visitor := NewDumpVisitor(ctx, "", nil)
	visitor.SetIndex(0)

	svc := &config.TCPService{
		Host:       "db.localhost",
		SSHBastion: "primary",
		TargetHost: "10.0.0.1",
		TargetPort: 5432,
		ListenPort: 5432,
	}

	err := visitor.VisitTCP(svc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configs := visitor.GetServiceConfigs()
	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}

	if configs[0].LocalPort != 10000 {
		t.Errorf("expected localPort 10000, got %d", configs[0].LocalPort)
	}
}
