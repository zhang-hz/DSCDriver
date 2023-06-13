export GIN_MODE=release
go env -w GOOS=linux
go env -w GOARCH=arm64
go build -o /mnt/c/data/code/daqdriverdist/dscv2_0 ./src