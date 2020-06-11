#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep TIME
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "PRIVKEY" | ./main -home /home/user/config -tss-port :8080 -peer /ip4/IPADDR/tcp/6668/ipfs/BOOTSTRAP

