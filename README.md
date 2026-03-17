# docker-port-scanner

`docker-port-scanner` は、Docker コンテナで公開されている TCP ポートを検出し、URL 形式で一覧表示する Go 製 CLI です。

## Features

- 引数で指定したコンテナ名またはコンテナ ID を対象にできる
- 引数未指定時は起動中の全コンテナを対象にできる
- 各コンテナの公開 TCP ポートを URL として表示する
- `--host` と `--scheme` で URL のホスト名とスキームを変更できる
- `Makefile` で主要 OS / ARCH 向けにクロスビルドできる

## Requirements

- Go 1.26 以上
- Docker CLI が利用可能であること
- Docker デーモンにアクセスできること

## Usage

単一コンテナを対象にする場合:

```bash
docker-port-scanner web
```

複数コンテナを対象にする場合:

```bash
docker-port-scanner web api
```

起動中の全コンテナを対象にする場合:

```bash
docker-port-scanner
```

ホスト名やスキームを変更する場合:

```bash
docker-port-scanner --host 127.0.0.1 --scheme https web
```

出力例:

```text
web
http://localhost:8080
http://localhost:8443

api
http://localhost:3000
```

## Build

単体でビルドする場合:

```bash
go build ./cmd/docker-port-scanner
```

主要 OS / ARCH 向けにクロスビルドする場合:

```bash
make build
```

成果物は `./build` に出力されます。

- `build/docker-port-scanner-linux-amd64`
- `build/docker-port-scanner-linux-arm64`
- `build/docker-port-scanner-darwin-amd64`
- `build/docker-port-scanner-darwin-arm64`
- `build/docker-port-scanner-windows-amd64.exe`
- `build/docker-port-scanner-windows-arm64.exe`

不要になった成果物は以下で削除できます。

```bash
make clean
```

## Development

テスト:

```bash
go test ./...
```

## Error Behavior

次のようなケースではエラーを返します。

- Docker デーモンに接続できない
- 指定したコンテナが存在しない
- 指定したコンテナが起動していない
- 指定したコンテナに公開 TCP ポートが存在しない
- 引数未指定で起動中コンテナが存在しない
- 引数未指定で、起動中コンテナに公開 TCP ポートを持つものが 1 件もない
