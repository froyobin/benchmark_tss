#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep 100
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "ODQ2NDg4MjQwMTAxNzE4ZTYyMWY5NzFhYjc3NWQzM2UzMTRkYjJlZWJlMmFkMDgwNjU2NDM1ZjMzNWQ2MzNhYg==" | ./main -home /home/user/config -tss-port :8080 -peer /ip4/IPADDR/tcp/6668/ipfs/16Uiu2HAm4cFVFafrrP18JrVG8WT9xHgPA18jVfiAHfTuzDBKwBe5

