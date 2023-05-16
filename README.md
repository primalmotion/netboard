# Netboard

Netboard is a client/server application that allows to share your clipboard
between different devices. It works by deploying a server, then running client
on your devices to sync the clipboard.

The server uses websockets or chunked http encoding to push data to client, and
while it is easier to handle the client part with the netboard client, it's
perfectly possible to integrate it with anything else, like good old curl.

> NOTE: netboard is NOT multi tenant. While it is possible to add it if there is
> any interest in the matter, for now it just assumes that anyone connecting
> with a valid certificate is willing to get its clipboard synchronized with the
> rest of the instance.

## Installation

### Getting the software

Using Go:

```
go install github.com/primalmotion/netboard@latest
```

The software is also packaged for Arch Linux and derivatives from the AUR:

```
yay -S netboard
```

Alternatively, if you want to use latest and greatest:

```
yay -S netboard-git
```

### Certificates

Netboard uses mutual TLS to authenticate the clients. This is not optional, and
you must generate several certificates:

- one certificate for the server (https)
- one certificate authority for the clients
- as many client certificates signed by the CA as you have devices.

You can generate certificates with the tool you like. This example will use
[tg](https://github.com/paloaltonetworks/tg).

First, we will generate the server certificate. You need to know in advance the
hostname or IP the server will use.

```sh
tg cert --name netboard-server --dns my.netboard.com --ip 127.0.0.1
```

> NOTE: You can sign that certificate with a CA you maintain, or get a
> certificate from let's encrypt.

Then we will create the CA needed to authenticate the users:

```sh
tg cert --is-ca --name netboard-client-ca
```

And finally client certificates:

```sh
tg cert \
  --signing-cert netboard-client-ca-cert.pem \
  --signing-cert-key netboard-client-ca-key.pem \
  --auth-client \
  --name my-laptop

tg cert \
  --signing-cert netboard-client-ca-cert.pem \
  --signing-cert-key netboard-client-ca-key.pem \
  --auth-client \
  --name my-phone
```

> NOTE: don't reuse the same certificate over several devices as the fingerprint
> will be used to prevent sending back data you just sent, resulting in a
> update larsen.

### Server

Copy over the `netboard-server-cert.pem` and `netboard-server-key.pem`, as well
as the `netboard-client-ca-cert.pem` to the machine hosting the server. We will
put them in `/etc/netboard`. Then write a `config.yaml` file with the following
content:

```yaml
server:
  cert: /etc/netboard/netboard-server-cert.pem
  key: /etc/netboard/netboard-server-key.pem
  client-ca: /etc/netboard/netboard-client-ca-cert.pem
```

Then run the server:

```sh
netboard server
```

### Clients

Copy the client certificate to the appropriate devices, in `~/.config/netboard`,
then write a `config.yaml` file with the following content:

```yaml
listen:
  url: https://my.netboard.com:8989
  cert: $HOME/.config/netboard/my-laptop-cert.pem
  cert-key: $HOME/.config/netboard/my-laptop-key.pem
```

Then run the client:

```sh
netboard listen
```

Repeat the process for every other clients you want to have all their clipboard
synced.


## Clipboard management modes

The netboard client can run using 2 modes, controlled by the `--mode` flag:

- `lib`: Uses native clipboard management libraries. This requires CGO and only
    works partially under Wayland
- `wl-clipboard`: Uses `wl-copy` and `wl-paste` utilities to handle the
    clipboard. This only works under Wayland.

More modes may come, as well as a smart way to choose the best for the current
platform.


## Clipboard listening mode

You can choose to use either chunked HTTP encoding (`--websocket false`) or
websockets (`--websocket true`, the default). Websocket should offer better and
faster pushes and connectivity loss detection.


## Usage


### Main command

```
$ netboard --help
Simple and secure network clipboard sharing engine

Usage:
  netboard [flags]
  netboard [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  listen      Sync data between clipboard and server
  server      Run the server

Flags:
  -h, --help      help for netboard
      --version   Show version

Use "netboard [command] --help" for more information about a command.
```

### Server command

```
$ netboard server --help
Run the server

Usage:
  netboard server [flags]

Flags:
  -c, --cert string            path to the server public key
  -k, --cert-key string        path to the server private key
  -p, --cert-key-pass string   optional server key passphrase
  -C, --client-ca string       path to the client certificate CA
  -h, --help                   help for server
  -l, --listen string          The listen address of the server (default ":8989")
```

### Listen command

```
$ netboard listen --help
Sync data between clipboard and server

Usage:
  netboard listen [flags]

Flags:
  -c, --cert string            Path to the client public key
  -k, --cert-key string        Path to the client private key
  -p, --cert-key-pass string   Optional client key passphrase
  -h, --help                   help for listen
      --insecure-skip-verify   Skip server CA validation. this is not secure
      --mode string            Select the mode to handle clipboard. wl-clipboard or lib (default "wl-clipboard")
  -C, --server-ca string       Path to the server certificate CA
  -u, --url string             The address of the netboard server (default "https://127.0.0.1:8989")
  -w, --websocket              Use websockets instead of chunked encoding (default true)
```
