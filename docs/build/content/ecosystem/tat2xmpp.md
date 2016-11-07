---
title: "tat2xmpp"
weight: 2
toc: true
prev: "/ecosystem/mail2tat"

---

## TAT configuration

```bash

[...]
# TAT 2 XMPP Configuration
TAT_TAT2XMPP_USERNAME=tat.system.jabber
TAT_TAT2XMPP_URL=http://tat2xmpp.your-domain
TAT_TAT2XMPP_KEY=a-key-used-by-tat2xmpp
[...]

# Running TAT Engine
./api
```

## TAT2XMPP Configuration

```bash
TAT2XMPP_LISTEN_PORT=8080
TAT2XMPP_HOOK_KEY=a-key-used-by-tat2xmpp
TAT2XMPP_USERNAME_TAT_ENGINE=tat.system.jabber
TAT2XMPP_XMPP_BOT_PASSWORD=password-of-bot-user-on-xmpp
TAT2XMPP_PRODUCTION=true
TAT2XMPP_PASSWORD_TAT_ENGINE=very-long-tat-password-of-tat.system.jabber
TAT2XMPP_XMPP_BOT_JID=robot.tat@your-domain
TAT2XMPP_XMPP_SERVER=your-jabber-server:5222
TAT2XMPP_URL_TAT_ENGINE=http://tat.your-domain

# Running TAT2XMPP
./tat2xmpp
```


## Usage

### Building

```bash
mkdir -p $GOPATH/src/github.com/ovh
cd $GOPATH/src/github.com/ovh
git clone git@github.com:ovh/tat-contrib.git
cd tat-contrib/tat2xmpp/api
go build
./api -h
```

### Flags

```bash

./api -h
Tat2XMPP

Usage:
  tat2xmpp [flags]
  tat2xmpp [command]

Available Commands:
  version     Print the version.

Flags:
  -c, --configFile string            configuration file
      --hook-key string              Hook Key, for using POST http://<url>/hook endpoint, with Header TAT2XMPPKEY
      --listen-port string           RunKPI Listen Port (default "8088")
      --log-level string             Log Level : debug, info or warn
      --password-tat-engine string   Password Tat Engine
      --production                   Production mode
      --url-tat-engine string        URL Tat Engine (default "http://localhost:8080")
      --username-tat-engine string   Username Tat Engine (default "tat.system.xmpp")
      --xmpp-bot-jid string          XMPP Bot JID (default "tat@localhost")
      --xmpp-bot-password string     XMPP Bot Password
      --xmpp-debug                   XMPP Debug
      --xmpp-insecure-skip-verify    XMPP InsecureSkipVerify (default true)
      --xmpp-notls                   XMPP No TLS (default true)
      --xmpp-server string           XMPP Server
      --xmpp-session                 XMPP Session (default true)
      --xmpp-starttls                XMPP Start TLS

Use "tat2xmpp [command] --help" for more information about a command.

```
