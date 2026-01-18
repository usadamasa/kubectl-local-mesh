package port

import (
	"fmt"
	"io"
	"net"
	"os"
)

// Port はすべてのポート番号型の共通制約
type Port interface {
	~int
}

// LocalPort はローカルポートフォワード用のポート番号
type LocalPort int

// ListenerPort はEnvoyメインリスナーのポート番号
type ListenerPort int

// ServicePort はKubernetesサービスのポート番号
type ServicePort int

// TCPPort はTCP接続のポート番号（target_port, listen_port）
type TCPPort int

// IndividualListenerPort は個別リスナーのポート番号
type IndividualListenerPort int

const (
	minPort           = 1
	maxPort           = 65535
	privilegedPortMax = 1023
)

// IsValid はポート番号が有効範囲内かを確認
func IsValid[T Port](p T) bool {
	return int(p) >= minPort && int(p) <= maxPort
}

// IsPrivileged は特権ポート（0-1023）かを確認
func IsPrivileged[T Port](p T) bool {
	return int(p) <= privilegedPortMax
}

// ValidatePort はポート番号のバリデーションを行う
func ValidatePort[T Port](port T, fieldName, serviceName string) error {
	if int(port) < minPort || int(port) > maxPort {
		return fmt.Errorf("%s must be between %d and %d for service '%s', got %d",
			fieldName, minPort, maxPort, serviceName, int(port))
	}
	return nil
}

// ValidateRequiredPort は必須ポート番号のバリデーションを行う
func ValidateRequiredPort[T Port](port T, fieldName, serviceName string) error {
	if int(port) == 0 {
		return fmt.Errorf("%s is required for service '%s'", fieldName, serviceName)
	}
	return ValidatePort(port, fieldName, serviceName)
}

// ValidatePorts はポート番号スライスのバリデーションを行う
func ValidatePorts[T Port](ports []T, fieldName, serviceName string) error {
	for _, p := range ports {
		if err := ValidatePort(p, fieldName, serviceName); err != nil {
			return err
		}
	}
	return nil
}

// warnWriter は警告出力先（テスト時に差し替え可能）
var warnWriter io.Writer = os.Stderr

// WarnWriter は現在の警告出力先を取得（テスト用）
func WarnWriter() io.Writer {
	return warnWriter
}

// SetWarnWriter は警告出力先を設定（テスト用）
func SetWarnWriter(w io.Writer) {
	warnWriter = w
}

// WarnPrivilegedPort は特権ポート使用時の警告を出力
func WarnPrivilegedPort[T Port](port T, fieldName, serviceName string) {
	p := int(port)
	if p > 0 && p <= privilegedPortMax {
		_, _ = fmt.Fprintf(warnWriter, "Warning: %s=%d for service '%s' is a privileged port (requires root/sudo)\n",
			fieldName, p, serviceName)
	}
}

// WarnLocalhostTLD はTCPサービスで.localhostドメインを使用した場合の警告を出力
// macOSでは.localhostサブドメインは/etc/hostsを無視して127.0.0.1に解決されるため、
// 異なるloopback IPにバインドされたTCPサービスに接続できない
func WarnLocalhostTLD(hostname, serviceName string) {
	// ".localhost"で終わるが、"localhost"そのものではない場合に警告
	if hostname != "localhost" && (hostname == ".localhost" || len(hostname) > 10 && hostname[len(hostname)-10:] == ".localhost") {
		_, _ = fmt.Fprintf(warnWriter, "Warning: TCP service '%s' uses .localhost domain (%s)\n"+
			"  macOS ignores /etc/hosts for .localhost subdomains and resolves them to 127.0.0.1\n"+
			"  This may cause connection failures. Consider using .localdomain or another TLD instead.\n",
			serviceName, hostname)
	}
}

// PortConflictChecker はポート競合を検出するためのヘルパー
type PortConflictChecker struct {
	usedPorts     map[int]string    // port -> service name（IP指定なし）
	usedAddrPorts map[string]string // "ip:port" -> service name（IP指定あり）
	wildcardPorts map[int]string    // 0.0.0.0でバインドされたport -> service name
}

// NewPortConflictChecker は新しいPortConflictCheckerを生成
func NewPortConflictChecker() *PortConflictChecker {
	return &PortConflictChecker{
		usedPorts:     make(map[int]string),
		usedAddrPorts: make(map[string]string),
		wildcardPorts: make(map[int]string),
	}
}

// Register はポートを登録し、競合があれば警告を出力（IP指定なし、従来互換）
func (c *PortConflictChecker) Register(port int, serviceName string) {
	if existingService, ok := c.usedPorts[port]; ok {
		_, _ = fmt.Fprintf(warnWriter, "Warning: port %d is used by both '%s' and '%s'\n",
			port, existingService, serviceName)
	}
	c.usedPorts[port] = serviceName
}

// RegisterWithAddr はIP:portの組み合わせで登録し、競合があれば警告を出力
// addr が空の場合はポート番号のみでチェック（従来動作）
// addr が "0.0.0.0" の場合は全インターフェースバインドとして扱う
func (c *PortConflictChecker) RegisterWithAddr(addr string, port int, serviceName string) {
	// アドレス指定がない場合は従来の動作
	if addr == "" {
		c.Register(port, serviceName)
		return
	}

	// 0.0.0.0（ワイルドカード）の場合
	if addr == "0.0.0.0" {
		// 既存のワイルドカードポートと競合チェック
		if existingService, ok := c.wildcardPorts[port]; ok {
			_, _ = fmt.Fprintf(warnWriter, "Warning: 0.0.0.0:%d is used by both '%s' and '%s'\n",
				port, existingService, serviceName)
		}
		// 既存の特定IPポートと競合チェック
		for key, existingService := range c.usedAddrPorts {
			existingPort := extractPort(key)
			if existingPort == port {
				_, _ = fmt.Fprintf(warnWriter, "Warning: port %d conflict - '%s' (0.0.0.0) binds all interfaces, conflicts with '%s' (%s)\n",
					port, serviceName, existingService, key)
			}
		}
		c.wildcardPorts[port] = serviceName
		return
	}

	// 特定IPアドレスの場合
	key := fmt.Sprintf("%s:%d", addr, port)

	// 既存のワイルドカードポートと競合チェック
	if existingService, ok := c.wildcardPorts[port]; ok {
		_, _ = fmt.Fprintf(warnWriter, "Warning: port %d conflict - '%s' (0.0.0.0) binds all interfaces, conflicts with '%s' (%s)\n",
			port, existingService, serviceName, key)
	}

	// 同じIP:portと競合チェック
	if existingService, ok := c.usedAddrPorts[key]; ok {
		_, _ = fmt.Fprintf(warnWriter, "Warning: %s is used by both '%s' and '%s'\n",
			key, existingService, serviceName)
	}
	c.usedAddrPorts[key] = serviceName
}

// extractPort は "ip:port" 形式の文字列からポート番号を抽出
func extractPort(addrPort string) int {
	_, portStr, err := net.SplitHostPort(addrPort)
	if err != nil {
		return 0
	}
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)
	return port
}

// RegisterPort はセマンティック型のポートを登録するジェネリック関数
func RegisterPort[T Port](c *PortConflictChecker, port T, serviceName string) {
	c.Register(int(port), serviceName)
}

// FreeLocalPort は利用可能なローカルポートを取得
func FreeLocalPort() (LocalPort, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return LocalPort(l.Addr().(*net.TCPAddr).Port), nil
}
