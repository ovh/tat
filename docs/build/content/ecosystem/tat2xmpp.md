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
