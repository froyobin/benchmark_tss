#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep 100
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "YTFlMTQ0MTU5YTZjOTg4MjU2OTg0ZGRmMzgwNDFmNThmNjMwMTkyNTZhYzNiNjgwOTQ0M2NmNTFkMDVkNDZlMA==" | ./main -home /home/user/config -tss-port :8080 -peer /ip4/128.199.106.202/tcp/6668/ipfs/16Uiu2HAm4cFVFafrrP18JrVG8WT9xHgPA18jVfiAHfTuzDBKwBe5

