package port

import (
	"fmt"
	"io"
	"net"
	"os"
)

// Port は汎用ポート番号の基底型
type Port int

// ListenerPort はEnvoyメインリスナーのポート番号
type ListenerPort int

// ServicePort はKubernetesサービスのポート番号
type ServicePort int

// TCPPort はTCP接続のポート番号（target_port, listen_port）
type TCPPort int

// IndividualListenerPort は個別リスナーのポート番号
type IndividualListenerPort int

const (
	MinPort           = 1
	MaxPort           = 65535
	PrivilegedPortMax = 1023
)

// IsValid はポート番号が有効範囲内かを確認
func (p Port) IsValid() bool {
	return p >= MinPort && p <= MaxPort
}

// IsPrivileged は特権ポート（0-1023）かを確認
func (p Port) IsPrivileged() bool {
	return p <= PrivilegedPortMax
}

// ValidatePort はポート番号のバリデーションを行う
func ValidatePort(port int, fieldName, serviceName string) error {
	if port < MinPort || port > MaxPort {
		return fmt.Errorf("%s must be between %d and %d for service '%s', got %d",
			fieldName, MinPort, MaxPort, serviceName, port)
	}
	return nil
}

// ValidateRequiredPort は必須ポート番号のバリデーションを行う
func ValidateRequiredPort(port int, fieldName, serviceName string) error {
	if port == 0 {
		return fmt.Errorf("%s is required for service '%s'", fieldName, serviceName)
	}
	return ValidatePort(port, fieldName, serviceName)
}

// ValidatePorts はポート番号スライスのバリデーションを行う
func ValidatePorts[T ~int](ports []T, fieldName, serviceName string) error {
	for _, p := range ports {
		if err := ValidatePort(int(p), fieldName, serviceName); err != nil {
			return err
		}
	}
	return nil
}

// warnWriter は警告出力先（テスト時に差し替え可能）
var warnWriter io.Writer = os.Stderr

// SetWarnWriter は警告出力先を設定（テスト用）
func SetWarnWriter(w io.Writer) {
	warnWriter = w
}

// WarnPrivilegedPort は特権ポート使用時の警告を出力
func WarnPrivilegedPort(port int, fieldName, serviceName string) {
	if port > 0 && port <= PrivilegedPortMax {
		_, _ = fmt.Fprintf(warnWriter, "Warning: %s=%d for service '%s' is a privileged port (requires root/sudo)\n",
			fieldName, port, serviceName)
	}
}

// PortConflictChecker はポート競合を検出するためのヘルパー
type PortConflictChecker struct {
	usedPorts map[int]string // port -> service name
}

// NewPortConflictChecker は新しいPortConflictCheckerを生成
func NewPortConflictChecker() *PortConflictChecker {
	return &PortConflictChecker{
		usedPorts: make(map[int]string),
	}
}

// Register はポートを登録し、競合があれば警告を出力
func (c *PortConflictChecker) Register(port int, serviceName string) {
	if existingService, ok := c.usedPorts[port]; ok {
		_, _ = fmt.Fprintf(warnWriter, "Warning: port %d is used by both '%s' and '%s'\n",
			port, existingService, serviceName)
	}
	c.usedPorts[port] = serviceName
}

// FreeLocalPort は利用可能なローカルポートを取得
func FreeLocalPort() (Port, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return Port(l.Addr().(*net.TCPAddr).Port), nil
}
