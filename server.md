# Running a skipctl server

1. `skipctl --output=json serve` with priviliges to do raw ICMP sockets.
2. Put it behind a reverse proxy terminating TLS (e.g. an Istio gateway or nginx)
3. Update DNS to make it available via service discovery from the clients

## DNS scheme

By default, the client will do a TXT lookup for the host defined as `DefaultDiscoveryServer` in [constants.go](./pkg/constants/constants.go) (optionally overridden by the client, `--discovery-host`).
For this document let's assume that it's `_skipctl.example.com`.

```shell
dig TXT _skipctl.example.com

; <<>> DiG 9.10.6 <<>> TXT _skipctl.example.com
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 44391
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 512
;; QUESTION SECTION:
;_skipctl.example.com.		IN	TXT

;; ANSWER SECTION:
_skipctl.example.com.	60	IN	TXT	"WwogICAgewogICAgICAgICJuYW1lIjogImF0a3YzLWV2ZW5oIiwKICAgICAgICAiYWRkciI6ICJsb2NhbGhvc3QuZXZlbmgubmV0OjQ0MyIKICAgIH0KXQ=="

[…]
```

The response is a base64 encoded JSON structure in the following form:
```json
[
  {
    "name": "myApiServer",
    "addr": "api-server-1.internal.example.com:443"
  }
]
```

In order to make more API servers available, this array must be expanded, base64 encoded and the TXT record must be updated. Be sure to validate both JSON and base64 encoding before updating the record.
