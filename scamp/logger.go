package scamp

import "os"
import "log"

var (
	Trace *log.Logger
	Info *log.Logger
	Warning *log.Logger
	Error *log.Logger
)

type NullWriter int
func (NullWriter) Write([]byte) (int, error) { return 0, nil }

func initSCAMPLogger() {
	// Idempotent logger setup!
	if Trace != nil {
		return;
	}

	Trace = log.New(new(NullWriter), "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

