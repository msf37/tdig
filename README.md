# tdig

build it:
```
go mod tidy
go build
```

basic lookup:
```
./tdig -domain google.com
```

tls lookup with cipher suite:
```
./tdig -domain google.com -suite TLS_AES_128_GCM_SHA256
```

tls lookup with cipher suite for specific dns server:
```
./tdig -domain google.com -server dns.google -suite TLS_AES_128_GCM_SHA256
```
