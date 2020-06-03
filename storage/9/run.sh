#!/bin/bash
echo "nameserver 8.8.8.8">>/etc/resolv.conf
sleep 100
go build ./cmd/tss/main.go ./cmd/tss/tss_http.go; echo "ZmU4NmJiOTExNzdlZjVlMDE0ODg2ZTVjMjkxNjYzN2YyZjBmNzAyNmMzMzBkOGNlMWI0ZGY3MWRmZTk3MzY4Ng==" | ./main -home /home/user/config -tss-port :8080 -peer /ip4/IPADDR/tcp/6668/ipfs/16Uiu2HAm4cFVFafrrP18JrVG8WT9xHgPA18jVfiAHfTuzDBKwBe5

