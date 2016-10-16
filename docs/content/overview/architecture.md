---
weight: 4
title: "Architecture"
prev: "/overview/lifecycle"
toc: true
---

![Architecture](/imgs/tat-architecture-overview.png?width=50%)

Main components of a Tat Installation are:

* Tat, also called Tat API or Tat Engine. Your can running as many Engine as you want.
You can imagine running many Tat instances on a Mesos / Marathon, or configure yourself behind a HAProxy.
* Tatwebui: it's a web application, an easier way to call Tat Engine than curl. This application
is component oriented, your can display messages with different Views. See [next chapter](/tatwebui) about Tatwebui for
more information.
* Tatcli, the TAT Command Line Interface. All Tat features are available on tatcli. You can use it
for "one shot" action on Tat API, or for display a UI in command line with `tatcli ui`. Tatcli ui
is popular to keep an eye on alerts or monitoring events without having a browser on it.
* uServices: Tat API is simple to be used, it's also easy to develop uService above tat to
make things a little more advanced, like on-call schedule & intervention, reporting, etc...
You'll find some opensourced uService in chapter [Ecosystem](/ecosystem)
