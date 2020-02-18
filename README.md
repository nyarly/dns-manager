# DNS Manager

The results of a coding exercise.

Using cobra for quick command line integration.

Single module for command line and service, to simplify serialization.

100 empty Zones.List() calls took 12s
(see ec1e291ebdb40fac063f0f01a5508109d81cd1e4)

Persistence: serving as a caching layer.
NS1 is considered authoritative, but we'd like faster responses and to shield our request counts

## Future Work

* TTL on cache
* Reduced interface (instead of transparent proxy of NS1)
* Input validation: A, MX and SRV records have different answer forms
* UI for multiple answers on a record
* Listing zones and records
