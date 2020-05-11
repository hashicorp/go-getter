module github.com/hashicorp/go-getter/gcs/v2

go 1.14

replace github.com/hashicorp/go-getter/v2 => ../

require (
	cloud.google.com/go/storage v1.6.0
	github.com/hashicorp/go-getter/v2 v2.0.0-00010101000000-000000000000
	google.golang.org/api v0.21.0
)
