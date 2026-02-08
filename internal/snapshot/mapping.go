package snapshot

import "github.com/usadamasa/kubectl-localmesh/internal/envoy"

// PortForwardMapping はポート割り当て結果を記録
type PortForwardMapping struct {
	Kind     string `yaml:"kind"`
	Host     string `yaml:"host"`
	Protocol string `yaml:"protocol,omitempty"`

	// Kubernetes service fields
	Namespace          string `yaml:"namespace,omitempty"`
	Service            string `yaml:"service,omitempty"`
	PortName           string `yaml:"port_name,omitempty"`
	Cluster            string `yaml:"cluster,omitempty"`
	ResolvedRemotePort int    `yaml:"resolved_remote_port,omitempty"`

	// TCP service fields
	SSHBastion string `yaml:"ssh_bastion,omitempty"`
	TargetHost string `yaml:"target_host,omitempty"`
	TargetPort int    `yaml:"target_port,omitempty"`

	// Common assigned ports
	AssignedLocalPort    int    `yaml:"assigned_local_port"`
	AssignedListenAddr   string `yaml:"assigned_listen_addr,omitempty"`   // TCPサービス用（loopback IP）
	AssignedListenerPort int    `yaml:"assigned_listener_port,omitempty"` // TCPサービス用 / 個別リスナーポート

	// Envoy cluster reference
	EnvoyClusterName string `yaml:"envoy_cluster_name"`
}

// PortForwardMappingSet はマッピングのセット
type PortForwardMappingSet struct {
	Services []PortForwardMapping `yaml:"services"`
}

// BuildMappings はServiceConfigのリストからPortForwardMappingSetを生成
func BuildMappings(configs []envoy.ServiceConfig) PortForwardMappingSet {
	mappings := make([]PortForwardMapping, 0, len(configs))

	for _, cfg := range configs {
		switch builder := cfg.Builder.(type) {
		case *envoy.KubernetesServiceBuilder:
			mapping := PortForwardMapping{
				Kind:               "kubernetes",
				Host:               builder.Host,
				Protocol:           builder.Protocol,
				Namespace:          builder.Namespace,
				Service:            builder.ServiceName,
				PortName:           builder.PortName,
				Cluster:            builder.Cluster,
				ResolvedRemotePort: int(cfg.ResolvedRemotePort),
				AssignedLocalPort:  int(cfg.LocalPort),
				EnvoyClusterName:   cfg.ClusterName,
			}

			// OverwriteListenPortがある場合はリスナーポートも記録
			if builder.OverwriteListenPort != 0 {
				mapping.AssignedListenerPort = int(builder.OverwriteListenPort)
			}

			mappings = append(mappings, mapping)

		case *envoy.TCPServiceBuilder:
			mapping := PortForwardMapping{
				Kind:                 "tcp",
				Host:                 builder.Host,
				SSHBastion:           builder.SSHBastion,
				TargetHost:           builder.TargetHost,
				TargetPort:           int(builder.TargetPort),
				AssignedLocalPort:    int(cfg.LocalPort),
				AssignedListenAddr:   builder.ListenAddr,
				AssignedListenerPort: int(builder.ListenPort),
				EnvoyClusterName:     cfg.ClusterName,
			}
			mappings = append(mappings, mapping)
		}
	}

	return PortForwardMappingSet{
		Services: mappings,
	}
}
