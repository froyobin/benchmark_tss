#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep 100
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "NjhlZWMzOTQ4ZjNjYmRlNWU1YTFjMzYyMTYwY2Q0YWU5NzBhZTA2ZTkwYmViYzhjNTI4YTc2MTdkMzg3NzhlNA==" | ./main -home /home/user/config -tss-port :8080 -peer /ip4/128.199.106.202/tcp/6668/ipfs/16Uiu2HAm4cFVFafrrP18JrVG8WT9xHgPA18jVfiAHfTuzDBKwBe5

