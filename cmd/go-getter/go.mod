module github.com/go-getter/cmd/go-getter/v2

go 1.14

replace (
	github.com/hashicorp/go-getter/gcs/v2 => ../../gcs
	github.com/hashicorp/go-getter/s3/v2 => ../../s3
	github.com/hashicorp/go-getter/v2 => ../..
)

require (
	github.com/cheggaaa/pb v1.0.28
	github.com/fatih/color v1.9.0 // indirect
	github.com/hashicorp/go-getter/gcs/v2 v2.0.0-00010101000000-000000000000
	github.com/hashicorp/go-getter/s3/v2 v2.0.0-00010101000000-000000000000
	github.com/hashicorp/go-getter/v2 v2.0.0-20201001102414-74576d5f550a
	github.com/mattn/go-runewidth v0.0.8 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
)
