module ivxv.ee/voting

go 1.23

require (
	github.com/gtank/ristretto255 v0.1.2 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	ivxv.ee/common/collector v1.9.11
	ivxv.ee/sessionstatus/api v1.9.11
)

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.17 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.17 // indirect
	go.etcd.io/etcd/client/v3 v3.5.17 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	tivi.io/core v0.2.2-0.20241127235149-149435f9e4f4
)

replace ivxv.ee/common/collector => ../common/collector

replace ivxv.ee/sessionstatus/api => ../sessionstatus/api
