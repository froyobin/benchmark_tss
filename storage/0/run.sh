#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep 1
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "YzQ1NjI5Zjc2MmVkNTBjY2M2ODFjYzExODNhNDhjYmMyOGUzMjkxZmE0M2QyZTY5ZTczMGIxMGJkZjAyZmM1OA==" | ./main -home /home/user/config -tss-port :8080 

