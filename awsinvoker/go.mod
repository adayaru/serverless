module github.com/ease-lab/vhive/examples/awsinvoker

go 1.19

replace (
	github.com/ease-lab/vhive/examples/endpoint => ../endpoint
)

require (
	github.com/ease-lab/vhive/examples/endpoint v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.44.88
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
