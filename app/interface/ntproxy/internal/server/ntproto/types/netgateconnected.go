package types

import (
)

type NtproNetgateConnected struct {
	NtproNetevent
}

var ConnectedEventId NtproEventtypeId
var DisconnectedEventId NtproEventtypeId

func init() {
	ConnectedEventId.Init(0, 1, 3, true)
	DisconnectedEventId.Init(0, 1, 4, true)
}





