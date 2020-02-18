# DNS Manager

The results of a coding exercise.

This is a simple tool to expose an NS1 DNS store and allow its zones and records to be manipulated.

You can install it by
```
go get github.com/nyarly/dns-manager
```
From there, you'll be able to invoke `dns-manager`. The command contains it's
own helpful documentation, but examples can be useful as well.

First of all, you'll need an API key from NS1. To keep the key from being
exposed e.g. via `ps`, `dns-manager` examines its environment to get the key - you'll need to
set `$NS1_APIKEY` to make it available. Once you've done that, you can
```
> dns-manager server
```
to start the service running. By default, it listens on "localhost:4444", but
that can be overriden with the `-listen` switch.

Once you have a server running, you can also use `dns-manager` to act as a client. In a separate terminal, you can try:
```
dns-manager zone add mynewzone.com
dns-manager record add www.mynewzone.com A 10.0.0.12
dns-manager record delete www.mynewzone A
dns-manager zone delete mynewzone.com
```

## Design notes

To stay within time contraints, the client was built as a command line
application, since there's a lot of solid tooling to support that in Go. With
that in mind, a single module for command line and service, simplified
serialization.

In discussion, it became clear that there was some necessary artificial
constraints, as befits an exercise. Specifically, NS1 publishes a very solid Go
package to interact with their service, so the server itself is a little superfluous.

To justify the project a little, the storage layer serves as a kind of cache,
to return results faster.  100 empty Zones.List() calls took 12s (see
ec1e291ebdb40fac063f0f01a5508109d81cd1e4), so this seemed like reasonable
utility to provide.

Further value was sought by building a more uniform interface for creating and
updating resources - both are perfectly reasonable uses of the PUT verb, so the
server dispatches as appropriate (based on its cache), while allowing clients
to be oblivious of the difference between existing and new records.

NS1 is considered authoritative, which limits how useful the cache can be. The
alternative, with `dns-manager` being the source of truth seemed ultimately
less teneble. Still, for reducing the amount of requests we make to NS1 to
retreive records, the cache is effective.

## Future Work

There are several things that I'd like to add with more time. Briefly enumerating a few of those:

* Listing zones and records
* Client support for viewing zones and records (individually and lists.)
* Cache records should expire after some reasonable TTL.
* The cache might also be used to decline re-creating Zones - it wasn't clear
  to me if deleted zones can ever be recreated, though.
* Rather than provide a transparent proxy of NS1, it might be worthwhile to
  reduce the representations provided to reflect the service's functionality
  better.
* Client side input validation: A, MX and SRV records have different answer forms
* UI for multiple answers on a record
* A more structured and performant persistence layer - a database most likely.
