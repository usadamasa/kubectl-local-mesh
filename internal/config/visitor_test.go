package config

import "testing"

type mockVisitor struct {
	visitedKubernetes bool
	visitedTCP        bool
}

func (m *mockVisitor) VisitKubernetes(*KubernetesService) error {
	m.visitedKubernetes = true
	return nil
}

func (m *mockVisitor) VisitTCP(*TCPService) error {
	m.visitedTCP = true
	return nil
}

func TestKubernetesService_Accept(t *testing.T) {
	svc := &KubernetesService{
		Host:      "test.localhost",
		Namespace: "default",
		Service:   "test-svc",
		Protocol:  "http",
	}

	visitor := &mockVisitor{}
	err := svc.Accept(visitor)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !visitor.visitedKubernetes {
		t.Error("expected VisitKubernetes to be called")
	}

	if visitor.visitedTCP {
		t.Error("expected VisitTCP not to be called")
	}
}

func TestTCPService_Accept(t *testing.T) {
	svc := &TCPService{
		Host:       "db.localhost",
		SSHBastion: "primary",
		TargetHost: "10.0.0.1",
		TargetPort: 5432,
		ListenPort: 5432,
	}

	visitor := &mockVisitor{}
	err := svc.Accept(visitor)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if visitor.visitedKubernetes {
		t.Error("expected VisitKubernetes not to be called")
	}

	if !visitor.visitedTCP {
		t.Error("expected VisitTCP to be called")
	}
}
