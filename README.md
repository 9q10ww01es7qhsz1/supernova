# supernova

Bloat-free DNS black hole server for personal use. It aims to be extremely simple while providing a bare minimum functionality to block ads.

```
GOOS=linux GOARCH=mips GOMIPS=softfloat go build -trimpath -ldflags="-s -w"
upx -9 supernova
```

```
curl -s https://v.firebog.net/hosts/AdguardDNS.txt -o black.list
./supernova
```