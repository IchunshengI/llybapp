PROTO_DIR := .
OUT_DIR := .
PATH := /root/go/bin:$(PATH)

.PHONY: all gen
all: gen

# trpc v1.0.9版本
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
