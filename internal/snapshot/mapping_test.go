package snapshot_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/usadamasa/kubectl-localmesh/internal/envoy"
	"github.com/usadamasa/kubectl-localmesh/internal/port"
	"github.com/usadamasa/kubectl-localmesh/internal/snapshot"
)

func TestPortForwardMapping_YAML(t *testing.T) {
	t.Run("Kubernetes service mapping", func(t *testing.T) {
		mapping := snapshot.PortForwardMapping{
			Kind:               "kubernetes",
			Host:               "api.localhost",
			Namespace:          "default",
			Service:            "api",
			Protocol:           "http",
			ResolvedRemotePort: 8080,
			AssignedLocalPort:  10000,
			EnvoyClusterName:   "default_api_8080",
		}

		// YAML出力確認
		b, err := yaml.Marshal(mapping)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		got := string(b)
		// 必須フィールドが含まれていることを確認
		assertContains(t, got, "kind: kubernetes")
		assertContains(t, got, "host: api.localhost")
		assertContains(t, got, "namespace: default")
		assertContains(t, got, "service: api")
		assertContains(t, got, "resolved_remote_port: 8080")
		assertContains(t, got, "assigned_local_port: 10000")
		assertContains(t, got, "envoy_cluster_name: default_api_8080")
	})

	t.Run("TCP service mapping", func(t *testing.T) {
		mapping := snapshot.PortForwardMapping{
			Kind:                 "tcp",
			Host:                 "db.localhost",
			SSHBastion:           "primary",
			TargetHost:           "10.0.0.1",
			TargetPort:           5432,
			AssignedLocalPort:    10001,
			AssignedListenAddr:   "127.0.0.2",
			AssignedListenerPort: 5432,
			EnvoyClusterName:     "tcp_primary_10_0_0_1_5432",
		}

		// YAML出力確認
		b, err := yaml.Marshal(mapping)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		got := string(b)
		// 必須フィールドが含まれていることを確認
		assertContains(t, got, "kind: tcp")
		assertContains(t, got, "host: db.localhost")
		assertContains(t, got, "ssh_bastion: primary")
		assertContains(t, got, "target_host: 10.0.0.1")
		assertContains(t, got, "target_port: 5432")
		assertContains(t, got, "assigned_local_port: 10001")
		assertContains(t, got, "assigned_listen_addr: 127.0.0.2")
		assertContains(t, got, "assigned_listener_port: 5432")
		assertContains(t, got, "envoy_cluster_name: tcp_primary_10_0_0_1_5432")
	})

	t.Run("omitempty works for optional fields", func(t *testing.T) {
		// Kubernetes serviceにはSSHBastion関連フィールドがない
		mapping := snapshot.PortForwardMapping{
			Kind:               "kubernetes",
			Host:               "api.localhost",
			Namespace:          "default",
			Service:            "api",
			Protocol:           "http",
			ResolvedRemotePort: 8080,
			AssignedLocalPort:  10000,
			EnvoyClusterName:   "default_api_8080",
		}

		b, err := yaml.Marshal(mapping)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		got := string(b)
		// SSHBastion関連フィールドは出力されないべき
		assertNotContains(t, got, "ssh_bastion")
		assertNotContains(t, got, "target_host")
		assertNotContains(t, got, "target_port")
		assertNotContains(t, got, "assigned_listen_addr")
		assertNotContains(t, got, "assigned_listener_port")
		// cluster未指定時は出力されないべき（envoy_cluster_nameとは別）
		assertNotContains(t, got, "\ncluster:")
	})
}

func TestPortForwardMappingSet_YAML(t *testing.T) {
	t.Run("multiple services", func(t *testing.T) {
		mappingSet := snapshot.PortForwardMappingSet{
			Services: []snapshot.PortForwardMapping{
				{
					Kind:               "kubernetes",
					Host:               "api.localhost",
					Namespace:          "default",
					Service:            "api",
					Protocol:           "http",
					ResolvedRemotePort: 8080,
					AssignedLocalPort:  10000,
					EnvoyClusterName:   "default_api_8080",
				},
				{
					Kind:                 "tcp",
					Host:                 "db.localhost",
					SSHBastion:           "primary",
					TargetHost:           "10.0.0.1",
					TargetPort:           5432,
					AssignedLocalPort:    10001,
					AssignedListenerPort: 5432,
					EnvoyClusterName:     "tcp_primary_10_0_0_1_5432",
				},
			},
		}

		b, err := yaml.Marshal(mappingSet)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		got := string(b)
		assertContains(t, got, "services:")
		assertContains(t, got, "kind: kubernetes")
		assertContains(t, got, "kind: tcp")
	})
}

func TestBuildMappings(t *testing.T) {
	t.Run("kubernetes services", func(t *testing.T) {
		builder := envoy.NewKubernetesServiceBuilder(
			"api.localhost", "http", "default", "api", "http", 0, 0, "",
		)
		configs := []envoy.ServiceConfig{
			{
				Builder:            builder,
				ClusterName:        "default_api_8080",
				LocalPort:          10000,
				ResolvedRemotePort: 8080,
			},
		}

		mappings := snapshot.BuildMappings(configs)

		if len(mappings.Services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(mappings.Services))
		}

		m := mappings.Services[0]
		assertEqual(t, "kubernetes", m.Kind)
		assertEqual(t, "api.localhost", m.Host)
		assertEqual(t, "default", m.Namespace)
		assertEqual(t, "api", m.Service)
		assertEqual(t, "http", m.Protocol)
		assertEqual(t, 8080, m.ResolvedRemotePort)
		assertEqual(t, 10000, m.AssignedLocalPort)
		assertEqual(t, "default_api_8080", m.EnvoyClusterName)
	})

	t.Run("kubernetes services with listener_port", func(t *testing.T) {
		builder := envoy.NewKubernetesServiceBuilder(
			"grpc.localhost", "grpc", "default", "grpc-service", "grpc", 0,
			port.IndividualListenerPort(8081),
			"",
		)
		configs := []envoy.ServiceConfig{
			{
				Builder:            builder,
				ClusterName:        "default_grpc_service_9090",
				LocalPort:          10001,
				ResolvedRemotePort: 9090,
			},
		}

		mappings := snapshot.BuildMappings(configs)

		if len(mappings.Services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(mappings.Services))
		}

		m := mappings.Services[0]
		assertEqual(t, "kubernetes", m.Kind)
		assertEqual(t, "grpc.localhost", m.Host)
		assertEqual(t, "grpc", m.Protocol)
		// OverwriteListenPortはAssignedListenerPortに記録
		assertEqual(t, 8081, m.AssignedListenerPort)
	})

	t.Run("tcp services", func(t *testing.T) {
		builder := envoy.NewTCPServiceBuilder(
			"db.localhost", 5432, "127.0.0.2", "primary", "10.0.0.1", 5432,
		)
		configs := []envoy.ServiceConfig{
			{
				Builder:     builder,
				ClusterName: "tcp_primary_10_0_0_1_5432",
				LocalPort:   10002,
			},
		}

		mappings := snapshot.BuildMappings(configs)

		if len(mappings.Services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(mappings.Services))
		}

		m := mappings.Services[0]
		assertEqual(t, "tcp", m.Kind)
		assertEqual(t, "db.localhost", m.Host)
		assertEqual(t, "primary", m.SSHBastion)
		assertEqual(t, "10.0.0.1", m.TargetHost)
		assertEqual(t, 5432, m.TargetPort)
		assertEqual(t, 10002, m.AssignedLocalPort)
		assertEqual(t, "127.0.0.2", m.AssignedListenAddr)
		assertEqual(t, 5432, m.AssignedListenerPort)
		assertEqual(t, "tcp_primary_10_0_0_1_5432", m.EnvoyClusterName)
	})

	t.Run("kubernetes services with cluster", func(t *testing.T) {
		builder := envoy.NewKubernetesServiceBuilder(
			"api.localhost", "http", "default", "api", "http", 0, 0,
			"gke_myproject_asia-northeast1_staging",
		)
		configs := []envoy.ServiceConfig{
			{
				Builder:            builder,
				ClusterName:        "default_api_8080",
				LocalPort:          10000,
				ResolvedRemotePort: 8080,
			},
		}

		mappings := snapshot.BuildMappings(configs)

		if len(mappings.Services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(mappings.Services))
		}

		m := mappings.Services[0]
		assertEqual(t, "kubernetes", m.Kind)
		assertEqual(t, "gke_myproject_asia-northeast1_staging", m.Cluster)
	})

	t.Run("mixed services", func(t *testing.T) {
		k8sBuilder := envoy.NewKubernetesServiceBuilder(
			"api.localhost", "http", "default", "api", "http", 0, 0, "",
		)
		tcpBuilder := envoy.NewTCPServiceBuilder(
			"db.localhost", 5432, "127.0.0.2", "primary", "10.0.0.1", 5432,
		)
		configs := []envoy.ServiceConfig{
			{
				Builder:            k8sBuilder,
				ClusterName:        "default_api_8080",
				LocalPort:          10000,
				ResolvedRemotePort: 8080,
			},
			{
				Builder:     tcpBuilder,
				ClusterName: "tcp_primary_10_0_0_1_5432",
				LocalPort:   10001,
			},
		}

		mappings := snapshot.BuildMappings(configs)

		if len(mappings.Services) != 2 {
			t.Fatalf("expected 2 services, got %d", len(mappings.Services))
		}

		// 順序を確認
		assertEqual(t, "kubernetes", mappings.Services[0].Kind)
		assertEqual(t, "tcp", mappings.Services[1].Kind)
	})
}

// Helper functions
func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !contains(got, want) {
		t.Errorf("expected to contain %q, got:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, notWant string) {
	t.Helper()
	if contains(got, notWant) {
		t.Errorf("expected NOT to contain %q, got:\n%s", notWant, got)
	}
}

func assertEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
