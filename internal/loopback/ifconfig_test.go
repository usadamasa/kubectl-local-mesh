package loopback

import (
	"testing"
)

func TestAliasManager_AddAlias(t *testing.T) {
	t.Run("ifconfigコマンドが正しい引数で呼ばれる", func(t *testing.T) {
		var calledArgs []string
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				calledArgs = args
				return nil
			},
		}

		err := mgr.AddAlias("127.0.0.2")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := []string{"lo0", "alias", "127.0.0.2", "up"}
		if len(calledArgs) != len(expected) {
			t.Errorf("expected %d args, got %d", len(expected), len(calledArgs))
		}
		for i, exp := range expected {
			if i < len(calledArgs) && calledArgs[i] != exp {
				t.Errorf("arg %d: expected %s, got %s", i, exp, calledArgs[i])
			}
		}
	})
}

func TestAliasManager_RemoveAlias(t *testing.T) {
	t.Run("ifconfigコマンドが正しい引数で呼ばれる", func(t *testing.T) {
		var calledArgs []string
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				calledArgs = args
				return nil
			},
		}

		err := mgr.RemoveAlias("127.0.0.2")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expected := []string{"lo0", "-alias", "127.0.0.2"}
		if len(calledArgs) != len(expected) {
			t.Errorf("expected %d args, got %d", len(expected), len(calledArgs))
		}
		for i, exp := range expected {
			if i < len(calledArgs) && calledArgs[i] != exp {
				t.Errorf("arg %d: expected %s, got %s", i, exp, calledArgs[i])
			}
		}
	})
}

func TestAliasManager_AddAliases(t *testing.T) {
	t.Run("複数のエイリアスを追加", func(t *testing.T) {
		var callCount int
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				callCount++
				return nil
			},
		}

		ips := []string{"127.0.0.2", "127.0.0.3", "127.0.0.4"}
		err := mgr.AddAliases(ips)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if callCount != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("空のリストでは何もしない", func(t *testing.T) {
		var callCount int
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				callCount++
				return nil
			},
		}

		err := mgr.AddAliases([]string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if callCount != 0 {
			t.Errorf("expected 0 calls, got %d", callCount)
		}
	})
}

func TestAliasManager_RemoveAliases(t *testing.T) {
	t.Run("複数のエイリアスを削除", func(t *testing.T) {
		var callCount int
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				callCount++
				return nil
			},
		}

		ips := []string{"127.0.0.2", "127.0.0.3", "127.0.0.4"}
		err := mgr.RemoveAliases(ips)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if callCount != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("削除エラーは無視して継続", func(t *testing.T) {
		var callCount int
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				callCount++
				// エラーを返しても継続する
				return &mockError{msg: "alias not found"}
			},
		}

		ips := []string{"127.0.0.2", "127.0.0.3"}
		// RemoveAliasesはエラーを無視する（クリーンアップ用途のため）
		err := mgr.RemoveAliases(ips)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestAliasManager_TracksAddedIPs(t *testing.T) {
	t.Run("AddAlias成功時にaddedに追加", func(t *testing.T) {
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				return nil
			},
			added: make([]string, 0),
		}

		_ = mgr.AddAlias("127.0.0.2")
		_ = mgr.AddAlias("127.0.0.3")

		added := mgr.GetAdded()
		if len(added) != 2 {
			t.Errorf("expected 2 added IPs, got %d", len(added))
		}
		if added[0] != "127.0.0.2" {
			t.Errorf("expected 127.0.0.2, got %s", added[0])
		}
		if added[1] != "127.0.0.3" {
			t.Errorf("expected 127.0.0.3, got %s", added[1])
		}
	})

	t.Run("AddAlias失敗時はaddedに追加しない", func(t *testing.T) {
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				return &mockError{msg: "permission denied"}
			},
			added: make([]string, 0),
		}

		_ = mgr.AddAlias("127.0.0.2")

		added := mgr.GetAdded()
		if len(added) != 0 {
			t.Errorf("expected 0 added IPs after failure, got %d", len(added))
		}
	})

	t.Run("RemoveAddedは追加成功分のみ削除", func(t *testing.T) {
		var removed []string
		mgr := &AliasManager{
			executor: func(name string, args ...string) error {
				// args = ["lo0", "-alias", "127.0.0.x"]
				if len(args) >= 3 && args[1] == "-alias" {
					removed = append(removed, args[2])
				}
				return nil
			},
			added: make([]string, 0),
		}

		_ = mgr.AddAlias("127.0.0.2")
		_ = mgr.AddAlias("127.0.0.3")

		mgr.RemoveAdded()

		if len(removed) != 2 {
			t.Errorf("expected 2 removed, got %d", len(removed))
		}
		if len(removed) >= 1 && removed[0] != "127.0.0.2" {
			t.Errorf("expected 127.0.0.2, got %s", removed[0])
		}
		if len(removed) >= 2 && removed[1] != "127.0.0.3" {
			t.Errorf("expected 127.0.0.3, got %s", removed[1])
		}

		// RemoveAdded後はaddedがクリアされていること
		if len(mgr.GetAdded()) != 0 {
			t.Errorf("expected added to be cleared after RemoveAdded, got %d", len(mgr.GetAdded()))
		}
	})

	t.Run("GetAddedはコピーを返す", func(t *testing.T) {
		mgr := &AliasManager{
			executor: func(name string, args ...string) error { return nil },
			added:    make([]string, 0),
		}

		_ = mgr.AddAlias("127.0.0.2")
		added := mgr.GetAdded()

		// 外部から変更しても内部には影響しない
		added[0] = "modified"
		internalAdded := mgr.GetAdded()
		if internalAdded[0] != "127.0.0.2" {
			t.Errorf("expected internal state to be unchanged, got %s", internalAdded[0])
		}
	})
}
