#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep 100
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "NjgwNGRlNTQ3ZDRjZGY5NWIxNWZhZTZhN2M1ZjQ4NzU3MzU5MWJmNzY2YmY5MWRlNzJmYWQ5NDQ2N2EzMjFkZQ==" | ./main -home /home/user/config -tss-port :8080 -peer /ip4/128.199.106.202/tcp/6668/ipfs/16Uiu2HAm4cFVFafrrP18JrVG8WT9xHgPA18jVfiAHfTuzDBKwBe5

