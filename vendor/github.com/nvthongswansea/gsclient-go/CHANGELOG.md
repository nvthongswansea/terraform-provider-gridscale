# Changelog

## 2.0.1 (Unreleased)

IMPROVEMENTS:
* Errors that are from http requests now include request UUIDs
* No need to create structs when exporting snapshots to S3

BUG FIXES:
* Fix README

## 2.0.0 (October 07, 2019)

FEATURES:
* Add `sync` mode
* The library now can be fully controlled through `context`
* Auto retry when server returns 5xx and 424 http codes
* Add a default configuration for the client

IMPROVEMENTS:
* Server type is now limited to pre-defined values
* Storage type is now limited to pre-defined values
* IP address family is now limited to pre-defined values
* Loadbalancer algorithm is now limited to pre-defined values
* All time-related properties are now type of GSTime (a custom type of time.Time)
* Friendly godoc.

BUG FIXES:
* Fixed bugs when unmarshalling JSON to concrete type (mismatched type)

## 1.0.0 (September 05, 2019)

FEATURES:
* Add support for Locations
* Add support for Events
* Add support for Labels
* Add support for Deletes

IMPROVEMENTS:
* Heavily code refactoring to improve code quality
* Achieve 95% test coverage
* Achieve 100% compliant golang code conventions based on goreportcard
* Power-off server if graceful shutdown fails
* Backward compatibility for server creation API


## 0.2.0 (August 23, 2019)

FEATURES:

* Add support for LBaaS (GH-2)
* Add support for PaaS (GH-6)
* Add support for ISO Image Handling (GH-8)
* Add support for Object Storage (GH-11)
* Add support for Snapshots (GH-12) and Snapshot Scheduler (GH-13)
* Add support for Firewall Handling (GH-14)

IMPROVEMENTS:

* Unit Tests for all functionality
* Logging support
* Many examples have been added
* Consistency as well as language styles improved

## 0.1.0 (January 2, 2019)

- Initial release of gsclient-go.

