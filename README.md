# mongo-watchdog

**Starting in Mongo 4.2 the
[Storage Node Watchdog](https://docs.mongodb.com/manual/administration/monitoring/#storage-node-watchdog)
is available in Community editions of Mongo and therefore this service is no
longer necessary and is no longer supported.**

MongoDB only supports a disk watchdog in the Enterprise edition. This small
binary can be run and will kill the local mongod if it remains unresponsive for
a set amount of time. It is based off the process described in
[SERVER-14139](https://jira.mongodb.org/browse/SERVER-14139?focusedCommentId=800618&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-800618).

Every `--update-interval` the binary tries to update a document in the `local`
DB and if it fails consecutively for more than `--failure-threshold` amount of
time the `mongod` process is killed, which should trigger the other secondaries
to perform an election.

## Installing

You can use go to build and install the binary by running:

```bash
go get github.com/levenlabs/mongo-watchdog
```

You can also download the latest release for your platform.

## Usage

```bash
./mongo-watchdog [--addr=127.0.0.1:27017] [--socket-timeout=30s] [--update-interval=15s] [--failure-threshold=2m]
```
