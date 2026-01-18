package loopback

import (
	"strings"
	"testing"
)

func TestIPAllocator_Allocate(t *testing.T) {
	t.Run("割り当てられたIPは127.0.0.x形式", func(t *testing.T) {
		alloc := NewIPAllocator()
		ip, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(ip, "127.0.0.") {
			t.Errorf("expected 127.0.0.x format, got %s", ip)
		}
	})

	t.Run("連続割り当てで全て異なるIPが返される", func(t *testing.T) {
		alloc := NewIPAllocator()
		seen := make(map[string]bool)
		for i := 0; i < 10; i++ {
			ip, err := alloc.Allocate()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if seen[ip] {
				t.Errorf("duplicate IP allocated: %s", ip)
			}
			seen[ip] = true
		}
	})

	t.Run("最大253個まで割り当て可能", func(t *testing.T) {
		alloc := NewIPAllocator()
		seen := make(map[string]bool)
		for i := 0; i < 253; i++ {
			ip, err := alloc.Allocate()
			if err != nil {
				t.Fatalf("unexpected error at %d: %v", i, err)
			}
			if seen[ip] {
				t.Errorf("duplicate IP allocated at %d: %s", i, ip)
			}
			seen[ip] = true
			if !strings.HasPrefix(ip, "127.0.0.") {
				t.Errorf("expected 127.0.0.x format, got %s", ip)
			}
		}
	})

	t.Run("順次割り当てで127.0.0.2から開始", func(t *testing.T) {
		alloc := NewIPAllocator()
		ip1, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ip2, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ip3, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ip1 != "127.0.0.2" {
			t.Errorf("expected 127.0.0.2, got %s", ip1)
		}
		if ip2 != "127.0.0.3" {
			t.Errorf("expected 127.0.0.3, got %s", ip2)
		}
		if ip3 != "127.0.0.4" {
			t.Errorf("expected 127.0.0.4, got %s", ip3)
		}
	})
}

func TestIPAllocator_Allocate_SkipsInUseIPs(t *testing.T) {
	t.Run("使用中IPをスキップ", func(t *testing.T) {
		alloc := NewIPAllocatorWithChecker(func(ip string) bool {
			// 127.0.0.2 と 127.0.0.4 は使用中として扱う
			return ip == "127.0.0.2" || ip == "127.0.0.4"
		})

		ip1, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ip1 != "127.0.0.3" {
			t.Errorf("expected 127.0.0.3, got %s", ip1)
		}

		ip2, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ip2 != "127.0.0.5" {
			t.Errorf("expected 127.0.0.5, got %s", ip2)
		}
	})

	t.Run("全IP使用中でエラー", func(t *testing.T) {
		alloc := NewIPAllocatorWithChecker(func(ip string) bool {
			return true // 全てのIPが使用中
		})

		_, err := alloc.Allocate()
		if err == nil {
			t.Error("expected error when all IPs are in use")
		}
	})

	t.Run("カスタムチェッカーなしでデフォルト動作", func(t *testing.T) {
		// NewIPAllocator() はデフォルトチェッカーを使用
		alloc := NewIPAllocator()
		ip, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(ip, "127.0.0.") {
			t.Errorf("expected 127.0.0.x format, got %s", ip)
		}
	})
}

func TestIPAllocator_GetAliases(t *testing.T) {
	t.Run("割り当てた全てのIPがエイリアスとして返される", func(t *testing.T) {
		alloc := NewIPAllocator()
		_, _ = alloc.Allocate()
		aliases := alloc.GetAliases()
		if len(aliases) != 1 {
			t.Errorf("expected 1 alias, got %d", len(aliases))
		}
	})

	t.Run("5つ割り当てた場合は5つのエイリアス", func(t *testing.T) {
		alloc := NewIPAllocator()
		for i := 0; i < 5; i++ {
			_, _ = alloc.Allocate()
		}
		aliases := alloc.GetAliases()
		if len(aliases) != 5 {
			t.Errorf("expected 5 aliases, got %d", len(aliases))
		}
		// 全てのエイリアスは127.0.0.x形式
		for _, alias := range aliases {
			if !strings.HasPrefix(alias, "127.0.0.") {
				t.Errorf("expected 127.0.0.x format, got %s", alias)
			}
		}
	})
}

func TestIPAllocator_Reset(t *testing.T) {
	t.Run("リセット後はエイリアスもクリア", func(t *testing.T) {
		alloc := NewIPAllocator()
		for i := 0; i < 5; i++ {
			_, _ = alloc.Allocate()
		}
		alloc.Reset()
		aliases := alloc.GetAliases()
		if len(aliases) != 0 {
			t.Errorf("expected 0 aliases after reset, got %d", len(aliases))
		}
	})
}
