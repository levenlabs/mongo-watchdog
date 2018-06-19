package main

import (
	"flag"
	"os"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	llog "github.com/levenlabs/go-llog"
)

type replInfo struct {
	IsMaster  bool
	Secondary bool
}

type serverStatus struct {
	PID  int
	Repl replInfo
}

func main() {
	st := flag.String("socket-timeout", "30s", "duration of the socket timeout")
	ui := flag.String("update-interval", "15s", "how often to try and update the local db")
	ft := flag.String("failure-threshold", "2m", "how long since the last successful update should we try to kill mongod")
	addr := flag.String("addr", "127.0.0.1:27017", "local mongo address, must be on the same server")
	flag.Parse()

	socketTimeout, err := time.ParseDuration(*st)
	if err != nil {
		llog.Fatal("failed to parse --socket-timeout", llog.ErrKV(err))
	}
	updateInterval, err := time.ParseDuration(*ui)
	if err != nil {
		llog.Fatal("failed to parse --update-interval", llog.ErrKV(err))
	}
	failureThreshold, err := time.ParseDuration(*ft)
	if err != nil {
		llog.Fatal("failed to parse --failure-threshold", llog.ErrKV(err))
	}
	sess, err := connect(*addr, socketTimeout)
	if err != nil {
		llog.Fatal("error connecting to mongo", llog.ErrKV(err))
	}
	spin(sess, updateInterval, failureThreshold)
}

func connect(addr string, timeout time.Duration) (*mgo.Session, error) {
	sess, err := mgo.DialWithTimeout("mongodb://"+addr+"/local?connect=direct", timeout)
	if err != nil {
		return nil, err
	}
	sess.SetSocketTimeout(timeout)
	sess.SetSafe(&mgo.Safe{
		J:        true,
		WTimeout: int(timeout.Nanoseconds() / 1e6),
	})
	sess.SetMode(mgo.Eventual, false)
	return sess, nil
}

func upsert(sess *mgo.Session) (serverStatus, error) {
	var status serverStatus
	if err := sess.Run("serverStatus", &status); err != nil {
		return status, err
	}
	if !status.Repl.IsMaster {
		return status, nil
	}
	// no need for this to be replicated so use the local database
	_, err := sess.DB("local").C("watchdog").Upsert(bson.M{
		"_id": "watchdog",
	}, bson.M{
		"lastUpdate": time.Now(),
	})
	return status, err
}

func spin(sess *mgo.Session, updateInterval, failureThreshold time.Duration) {
	var firstFailure time.Time

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()
	for range ticker.C {
		status, err := upsert(sess)
		if err == nil {
			firstFailure = time.Time{}
			continue
		}
		if status.PID == 0 {
			firstFailure = time.Time{}
			llog.Error("error running serverStatus", llog.ErrKV(err))
			// without a PID we can't do anything, so ignore
			continue
		}
		llog.Error("error updating local watchdog serverStatus", llog.ErrKV(err))
		if firstFailure.IsZero() {
			firstFailure = time.Now()
		}

		// if we're over the threshold, kill the process
		if time.Now().Sub(firstFailure) >= failureThreshold {
			llog.Info("killing mongod instance", llog.KV{"pid": status.PID})
			proc, err := os.FindProcess(status.PID)
			if err == nil {
				err = proc.Kill()
			}
			if err != nil {
				llog.Error("error finding and killing the mongod pid", llog.KV{"pid": status.PID}, llog.ErrKV(err))
			}
		}
	}
}
