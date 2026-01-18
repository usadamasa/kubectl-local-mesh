package loopback

import (
	"os/exec"
)

// CommandExecutor はコマンド実行の抽象化（テスト用）
type CommandExecutor func(name string, args ...string) error

// defaultExecutor は実際のifconfigコマンドを実行する
func defaultExecutor(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// AliasManager はloopback IPエイリアスの追加・削除を管理
type AliasManager struct {
	executor CommandExecutor
	added    []string // 追加成功したIPを追跡
}

// NewAliasManager は新しいAliasManagerを生成
func NewAliasManager() *AliasManager {
	return &AliasManager{
		executor: defaultExecutor,
		added:    make([]string, 0),
	}
}

// AddAlias は指定されたIPをlo0のエイリアスとして追加
// 成功した場合のみ内部で追跡し、RemoveAddedで削除可能にする
// sudo ifconfig lo0 alias <ip> up
func (m *AliasManager) AddAlias(ip string) error {
	if err := m.executor("ifconfig", "lo0", "alias", ip, "up"); err != nil {
		return err
	}
	m.added = append(m.added, ip)
	return nil
}

// GetAdded は追加成功したIPのリストをコピーとして返す
func (m *AliasManager) GetAdded() []string {
	result := make([]string, len(m.added))
	copy(result, m.added)
	return result
}

// RemoveAdded は追加成功したIPを全て削除する
// 削除時のエラーは無視する（クリーンアップ用途のため）
func (m *AliasManager) RemoveAdded() {
	for _, ip := range m.added {
		_ = m.executor("ifconfig", "lo0", "-alias", ip)
	}
	m.added = nil
}

// RemoveAlias は指定されたIPをlo0のエイリアスから削除
// sudo ifconfig lo0 -alias <ip>
func (m *AliasManager) RemoveAlias(ip string) error {
	return m.executor("ifconfig", "lo0", "-alias", ip)
}

// AddAliases は複数のIPをエイリアスとして追加
func (m *AliasManager) AddAliases(ips []string) error {
	for _, ip := range ips {
		if err := m.AddAlias(ip); err != nil {
			return err
		}
	}
	return nil
}

// RemoveAliases は複数のIPをエイリアスから削除
// エラーは無視して継続する（クリーンアップ用途のため）
func (m *AliasManager) RemoveAliases(ips []string) error {
	for _, ip := range ips {
		// エラーは無視（エイリアスが存在しない場合など）
		_ = m.RemoveAlias(ip)
	}
	return nil
}
