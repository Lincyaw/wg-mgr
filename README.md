
# Debian

Prepare:

```bash
sudo apt install wireguard
```

Usage:

1. Update server config

```bash
./vpn-tool setup > /etc/wireguard/wg0.conf

wg syncconf wg0 <(wg-quick strip wg0) # if already have installed interface

wg-quick up wg0 # else
``

2. Add user

```bash
./vpn-tool adduser --id yourname
```

3. Delete user

```bash
./vpn-tool deluser --id yourname
```

4. Delete user

```bash
./vpn-tool getuser --id yourname
```
