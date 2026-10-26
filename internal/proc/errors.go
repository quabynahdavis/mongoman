package proc

import "errors"

// errProcessNotFound is returned when no mongod process can be located
// for the given instance/port.
var errProcessNotFound = errors.New("mongod process not found")
