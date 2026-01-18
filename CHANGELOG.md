# Changelog

## [v0.3.2](https://github.com/usadamasa/kubectl-localmesh/compare/v0.3.1...v0.3.2) - 2026-01-18
### New Features 🎉
- feat: loopback IPエイリアスによるTCPサービスの同一ポート対応 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/66

## [v0.3.1](https://github.com/usadamasa/kubectl-localmesh/compare/v0.3.0...v0.3.1) - 2026-01-18
### Bug Fixes 🐛
- fix: Envoy domainsにhost:port形式を追加してgRPCクライアント互換性を改善 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/62
### Other Changes
- refactor: Portをジェネリック型制約に変更しキャスト削減 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/64

## [v0.3.0](https://github.com/usadamasa/kubectl-localmesh/compare/v0.2.1...v0.3.0) - 2026-01-17
### New Features 🎉
- feat: support overwrite_listen_port by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/56
- feat: ログレベル階層化とユーザーフレンドリーなサマリー出力を実装 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/59
### Other Changes
- Refactor/switch kind by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/53
- feat: add CLI-based snapshot testing for Envoy configuration by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/55
- refactor: ポート番号にセマンティック型を導入 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/57
- refactor: dump/snapshotパッケージを分離して責務を明確化 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/58
- feat: Envoy警告抑止とGCP SSH tunnel IAP明示指定を追加 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/60

## [v0.2.1](https://github.com/usadamasa/kubectl-localmesh/compare/v0.2.0...v0.2.1) - 2026-01-11
### Bug Fixes 🐛
- bugfix: suport http1 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/51
### Other Changes
- chore: add log by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/50

## [v0.2.0](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.7...v0.2.0) - 2026-01-03
### New Features 🎉
- support db via bastion by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/46

## [v0.1.7](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.6...v0.1.7) - 2025-12-30
### New Features 🎉
- feat: introduce Cobra-based subcommand structure with 'up' command by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/42
- refactor: reorganize CLI options and introduce dump-envoy-config subcommand by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/45

## [v0.1.6](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.5...v0.1.6) - 2025-12-30
### Bug Fixes 🐛
- Bugfix/handle invalid hosts by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/41
### Other Changes
- migrate to kubernetes/client-go from kubectl by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/36

## [v0.1.5](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.4...v0.1.5) - 2025-12-29
### Breaking Changes 🛠
- refactor: rename project from kubectl-local-mesh to kubectl-localmesh by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/33

## [v0.1.4](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.3...v0.1.4) - 2025-12-29
### Other Changes
- now by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/30

## [v0.1.3](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.2...v0.1.3) - 2025-12-29
### Bug Fixes 🐛
- fix: /etc/hostsの空行累積問題を修正 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/28
- adopt kubectl plugin naming by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/29

## [v0.1.2](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.1...v0.1.2) - 2025-12-29
### Bug Fixes 🐛
- bugfix: fix with golangci-lint by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/17

## [v0.1.1](https://github.com/usadamasa/kubectl-localmesh/compare/v0.1.0...v0.1.1) - 2025-12-28
- setup ci by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/11
- introduce tagpr by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/12
- run tagpr with gh app token by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/14

## [v0.1.0](https://github.com/usadamasa/kubectl-localmesh/commits/v0.1.0) - 2025-12-27
- [from now] 2025/12/27 17:58:16 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/2
- [from now] 2025/12/27 21:59:49 by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/3
- Make --update-hosts default to true for normal startup by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/4
- Change default listen port by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/6
- add ci-status-check by @usadamasa in https://github.com/usadamasa/kubectl-localmesh/pull/7
