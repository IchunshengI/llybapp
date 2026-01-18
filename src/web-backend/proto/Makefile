PROTO_DIR := .
OUT_DIR := .
PATH := /root/go/bin:$(PATH)

.PHONY: all gen
all: gen

gen:
	trpc create \
		--rpconly \
		--alias \
		--lang go \
		--protocol trpc \
		--protofile admin.proto \
		--protodir $(PROTO_DIR) \
		--output $(OUT_DIR) \
		--nogomod \
		--mock=false \
		--force
