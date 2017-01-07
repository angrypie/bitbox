all: golang
	go install

golang:
	protoc --go_out=plugins=grpc:. bitbox.proto
