# NetBoard

NetBoard is a client/server application that allows to share your clipboard
between different devices. It works by deploying a server, then running client
on your devices to sync the clipboard.

The server uses chunked http encoding to push data to client, and while it is
easier to handle the client part with the netboard client, it's perfectly
possible to integrate it with anything else, like good old curl.

## Installation

### Getting the software

Using Go:

```
go install git.sr.ht/~primalmotion/netboard@latest
```

The software is also packaged for Arch Linux and derivatives from the AUR:

```
yay -S netboard-git
```

### Certificates

Netboard uses mutual tls to authenticate the clients. This is not optional, and
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
put them in /etc/netboard. Then write a `config.yaml` file with the following
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

Copy the client certificate to the approriate devices, in ~/.config/netboard,
then write a config.yaml file with the following content:

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

## Modes

The netboard client can run using 2 modes, controlled by the `--mode` flag:

- `lib`: Uses native clipboard management libraries. This requires CGO and only
    works partially under Wayland
- `wl-clipboard`: Uses `wl-copy` and `wl-paste` utilities to handle the
    clipboard. This only works under Wayland.

More modes may come, as well as a smart way to choose the best for the current
platform.
