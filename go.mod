module github.com/terraform-providers/terraform-provider-gridscale

require (
	github.com/google/uuid v1.1.1 // indirect
	github.com/gridscale/gsclient-go v0.0.0-20190918114344-263938cc25e5
	github.com/hashicorp/terraform v0.12.5
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7 // indirect
	golang.org/x/sys v0.0.0-20190916202348-b4ddaad3f8a3 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.13
