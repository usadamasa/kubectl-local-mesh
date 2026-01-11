package config

// ServiceVisitor はすべてのサービス種別に対する操作を定義
type ServiceVisitor interface {
	// VisitKubernetes は Kubernetes Service に対する処理
	VisitKubernetes(*KubernetesService) error

	// VisitTCP は TCP Service に対する処理
	VisitTCP(*TCPService) error
}
