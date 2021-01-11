module github.com/The-New-Fork/email-pipeline

go 1.15

require (
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/emersion/go-imap v1.0.6
	github.com/go-chi/render v1.0.1
	github.com/jmoiron/jsonq v0.0.0-20150511023944-e874b168d07e
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	github.com/unchain/pipeline v0.0.0-20210111150105-791be2360b50
	github.com/unchainio/interfaces v0.2.1
	github.com/unchainio/pkg v0.22.1
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

replace github.com/spf13/viper v1.2.2 => github.com/unchainio/viper v1.2.2-0.20190712174521-9bf201c29832
