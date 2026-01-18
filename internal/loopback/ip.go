// Package loopback は、TCPサービス用のloopback IPエイリアス管理を提供する
// macOS限定機能（ifconfig lo0 aliasを使用）
package loopback

import (
	"errors"
	"fmt"
	"net"
)

// IPChecker はIPアドレスが使用中かを判定する関数型
type IPChecker func(ip string) bool

// IPAllocator は127.0.0.x系のIPアドレスを順次割り当てる
// 127.0.0.1は常に使用中のため除外し、127.0.0.2〜127.0.0.254から採番する
type IPAllocator struct {
	nextOctet int       // 次に割り当てるオクテット（2〜254）
	allocated []string  // 割り当て済みIPリスト
	isInUse   IPChecker // 使用中IPチェッカー
}

// NewIPAllocator は新しいIPAllocatorを生成（デフォルトチェッカー使用）
func NewIPAllocator() *IPAllocator {
	return NewIPAllocatorWithChecker(defaultIPChecker)
}

// NewIPAllocatorWithChecker はカスタムチェッカーでIPAllocatorを生成（テスト用）
func NewIPAllocatorWithChecker(checker IPChecker) *IPAllocator {
	return &IPAllocator{
		nextOctet: 2, // 127.0.0.2から開始
		allocated: make([]string, 0),
		isInUse:   checker,
	}
}

// defaultIPChecker はlo0インターフェースの既存IPをチェック
func defaultIPChecker(ip string) bool {
	iface, err := net.InterfaceByName("lo0")
	if err != nil {
		return false
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.String() == ip {
				return true
			}
		}
	}
	return false
}

// Allocate は次の利用可能なloopback IPを返す
// 使用中IPをスキップし、利用可能なIPがない場合はエラーを返す
func (a *IPAllocator) Allocate() (string, error) {
	for a.nextOctet <= 254 {
		ip := fmt.Sprintf("127.0.0.%d", a.nextOctet)
		a.nextOctet++
		if !a.isInUse(ip) {
			a.allocated = append(a.allocated, ip)
			return ip, nil
		}
	}
	return "", errors.New("loopback IP range exhausted (127.0.0.2-254)")
}

// GetAliases は追加が必要なエイリアスIPのリストを返す
// 127.0.0.2以降を使用するため、全てがエイリアス対象
func (a *IPAllocator) GetAliases() []string {
	return a.allocated
}

// Reset はアロケータをリセットして再利用可能にする
func (a *IPAllocator) Reset() {
	a.nextOctet = 2
	a.allocated = make([]string, 0)
}
