// Package log はkubectl-localmesh用のログ機能を提供します。
// ログレベルの階層化とユーザーフレンドリーな出力をサポートします。
package log

import (
	"fmt"
	"io"
	"os"
)

// Level はログレベルを表す定数です。
type Level int

const (
	// LevelWarn はエラーのみ出力する静かモード
	LevelWarn Level = iota
	// LevelInfo はデフォルトのログレベル（接続サマリー + 基本状態）
	LevelInfo
	// LevelDebug は詳細調査用のログレベル（再接続ログ + Envoy詳細）
	LevelDebug
)

// Logger はログ出力を管理する構造体です。
type Logger struct {
	level  Level
	writer io.Writer
}

// New は指定されたログレベルでLoggerを生成します。
// 出力先は標準出力になります。
func New(level string) *Logger {
	return NewWithWriter(level, os.Stdout)
}

// NewWithWriter は指定されたログレベルとWriterでLoggerを生成します。
// テスト用にWriterを注入できます。
func NewWithWriter(level string, w io.Writer) *Logger {
	return &Logger{
		level:  parseLevel(level),
		writer: w,
	}
}

// parseLevel は文字列からLogLevelに変換します。
func parseLevel(s string) Level {
	switch s {
	case "warn":
		return LevelWarn
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	default:
		// 未知のレベルはinfoとして扱う
		return LevelInfo
	}
}

// Level は現在のログレベルを文字列で返します。
func (l *Logger) Level() string {
	switch l.level {
	case LevelWarn:
		return "warn"
	case LevelDebug:
		return "debug"
	default:
		return "info"
	}
}

// EnvoyLevel はEnvoyに渡すログレベルを返します。
// infoレベルの場合はwarnを返し、Envoyの詳細ログを抑制します。
func (l *Logger) EnvoyLevel() string {
	switch l.level {
	case LevelDebug:
		return "debug"
	default:
		// info, warnの場合はEnvoyをwarnに設定して静かにする
		return "warn"
	}
}

// ShouldLogInfo はinfoレベルのログを出力すべきかどうかを返します。
func (l *Logger) ShouldLogInfo() bool {
	return l.level >= LevelInfo
}

// ShouldLogDebug はdebugレベルのログを出力すべきかどうかを返します。
func (l *Logger) ShouldLogDebug() bool {
	return l.level >= LevelDebug
}

// Info はinfoレベルのメッセージを出力します。
func (l *Logger) Info(msg string) {
	if l.ShouldLogInfo() {
		_, _ = fmt.Fprintln(l.writer, msg)
	}
}

// Infof はinfoレベルのフォーマット付きメッセージを出力します。
func (l *Logger) Infof(format string, args ...any) {
	if l.ShouldLogInfo() {
		_, _ = fmt.Fprintf(l.writer, format+"\n", args...)
	}
}

// Debug はdebugレベルのメッセージを出力します。
func (l *Logger) Debug(msg string) {
	if l.ShouldLogDebug() {
		_, _ = fmt.Fprintf(l.writer, "[DEBUG] %s\n", msg)
	}
}

// Debugf はdebugレベルのフォーマット付きメッセージを出力します。
func (l *Logger) Debugf(format string, args ...any) {
	if l.ShouldLogDebug() {
		_, _ = fmt.Fprintf(l.writer, "[DEBUG] "+format+"\n", args...)
	}
}
