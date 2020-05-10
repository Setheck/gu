package health

import (
	"net/http"
	"time"
)

type Reporter interface {
	Initialize() error
	Health() Event
	SetStdOutFallback(b bool)
	SetErrFn(efn func(err error))
	SetHealth(state State, message string)
	PersistStat(key string, val interface{})
	RegisterStatFn(name string, fn func(reporter Reporter))
	AddStat(key string, val interface{})
	GetStat(key string) interface{}
	ClearStat(key string)
	ClearStats()
	ClearStatFns()
	ClearAll()
	ReportHealth()
	AddHostnameSuffix(string)
	StartIntervalReporting(interval time.Duration)
	HealthHandler(w http.ResponseWriter, req *http.Request)
	Stop() error
	StopWithFinalState(final State, msg string) error
}

type Producer interface {
	Produce(message []byte) error
	ProducerHealth() bool
}
