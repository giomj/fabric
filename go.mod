module github.com/openziti/fabric

go 1.14

// replace github.com/openziti/foundation => ../foundation

require (
	github.com/alangpierce/go-forceexport v0.0.0-20160317203124-8f1d6941cd75 // indirect
	github.com/emirpasic/gods v1.12.0
	github.com/golang/protobuf v1.4.0
	github.com/google/go-cmp v0.5.1
	github.com/google/uuid v1.1.1
	github.com/marten-seemann/chacha20 v0.2.0 // indirect
	github.com/michaelquigley/pfxlog v0.0.0-20190813191113-2be43bd0dccc
	github.com/openziti/foundation v0.11.17
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	go.etcd.io/bbolt v1.3.5-0.20200615073812-232d8fc87f50
	gopkg.in/yaml.v2 v2.3.0
)
