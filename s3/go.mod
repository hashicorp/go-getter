module github.com/hashicorp/go-getter/s3/v2

go 1.14

replace github.com/hashicorp/go-getter/v2 => ../

require (
	github.com/aws/aws-sdk-go v1.44.114
	github.com/hashicorp/go-getter/v2 v2.1.1
)
