package main

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const introspectXML = `
<node>
  <interface name="com.subgraph.sublogmon">
    <method name="Logger">
      <arg name="id" direction="in" type="s" />
      <arg name="level" direction="in" type="u" />
      <arg name="timestamp" direction="in" type="u" />
      <arg name="logline" direction="in" type="s" />
    </method>
  </interface>` +
	introspect.IntrospectDataString +
	`</node>`

type dbusObject struct {
	dbus.BusObject
}

type dbusServer struct {
	conn *dbus.Conn
	run  bool
}

type slmData struct {
	EventID     string
	LogLevel    string
	Timestamp   int64
	LogLine     string
	OrigLogLine string
	Metadata    map[string]string
}

const busName = "com.subgraph.sublogmon"
const objectPath = "/com/subgraph/sublogmon"
const interfaceName = "com.subgraph.sublogmon"

func newDbusObject() (*dbusObject, error) {
	conn, err := dbus.SystemBus()

	if err != nil {
		return nil, err
	}

	return &dbusObject{conn.Object("com.subgraph.EventNotifier", "/com/subgraph/EventNotifier")}, nil
}

func (ob *dbusObject) alertObj(id, level string, timestamp int64, line, oline string, metadata map[string]string) {
	dobj := slmData{id, level, timestamp, line, oline, metadata}
	ob.Call("com.subgraph.EventNotifier.Alert", 0, dobj)
}

func newDbusServer() (*dbusServer, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return nil, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, errors.New("Bus name is already owned")
	}
	ds := &dbusServer{conn: conn, run: true}

	if err := conn.Export(ds, objectPath, interfaceName); err != nil {
		return nil, err
	}
	if err := conn.Export(introspect.Introspectable(introspectXML), objectPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *dbusServer) Logger(id string, level uint16, timestamp uint64, logline string) (bool, *dbus.Error) {
	fmt.Printf("<- Logger(id=%s, level=%v, timestamp=%v, line = [%s])\n", id, level, timestamp, logline)
	globalDBO.alertObj(id, string(level), int64(timestamp), logline, logline, nil)
	return true, nil
}
