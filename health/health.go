package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Setheck/gu"
	"github.com/Setheck/gu/build"
	"github.com/Setheck/gu/util"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const MaxHealthCheckInterval = time.Minute * 5

// State represents a visual indicator of the health
type State string

const (
	Blue   = State("blue")   // init or shut-down
	Green  = State("green")  // A-OK
	Yellow = State("yellow") // may need assistance
	Red    = State("red")    // HELP!
	Gray   = State("gray")   // maintenance mode
)

// Set compiles with the Flag.Value interface
func (s *State) Set(h string) error {
	switch h {
	case string(Blue):
		*s = Blue
	case string(Green):
		*s = Green
	case string(Yellow):
		*s = Yellow
	case string(Red):
		*s = Blue
	case string(Gray):
		*s = Gray
	default:
		return errors.New(fmt.Sprintf("unknown State: %s", h))
	}

	return nil
}

// Set compiles with the Flag.Value interface (and Stringer)
func (s State) String() string {
	return string(s)
}

// StateDefault health is Blue
const StateDefault = Blue

// EventType value is "health"
const EventType = "health"

// Event is the basic format for all JSON-encoded Health
// Example of a json encoded health event
//	{"hostname":"host123", "timestamp":1559761560, "etype":"health", "event":"status",
//	 "service":"gather", "version":"v0.0.1", "state":"green", "msg":"a-ok",
//	 "data": { "testmode": "off", "kernel": true }}
type Event struct {
	Hostname  string `json:"hostname"`  // Emitting host
	Timestamp int64  `json:"timestamp"` // Timestamp of event
	Type      string `json:"etype"`     // Type of event ( "health")
	Name      string `json:"event"`     // Name of event (e.g. status, startup, shutdown)
	Service   string `json:"service"`   // Operating service (e.g. gather)
	Version   string `json:"version"`   // Service version
	State     State  `json:"state"`     // Enumeration of health
	Message   string `json:"msg"`       // User actionable message

	Data interface{} `json:"data,omitempty"` // Service-specific data
}

type VersionMsg struct {
	Version            string `json:"version"`
	Build              string `json:"build"`
	Hostname           string `json:"hostname"`
	Kernel             string `json:"kernel"`
	ModuleLoaded       bool   `json:"loaded"`
	Timestamp          int64  `json:"timestamp"`
	LastEventTimestamp int64  `json:"last"`
	Count              uint64 `json:"count"`
	EventsSkipped      int    `json:"skipped"`
	Health             State  `json:"health"`
}

type HealthMsg struct {
	Hostname  string `json:"hostname"`
	Timestamp int64  `json:"timestamp"`
	Etype     string `json:"etype"`
	Event     string `json:"event"`

	Build string `json:"build,omitempty"`
}

//var HealthTopicSuffix = "system.health"

func GetHostHealthBytes(now time.Time) []byte {
	v, _ := mem.VirtualMemory()

	parts, err := disk.Partitions(false)
	check(err)

	usage := make(map[string]*disk.UsageStat)

	for _, part := range parts {
		u, err := disk.Usage(part.Mountpoint)
		check(err)
		usage[part.Mountpoint] = u
	}
	hostname, _ := os.Hostname()

	// cpu - get CPU number of cores and speed
	//cpuStat, err := cpu.Info()
	//check(err)

	times, err := cpu.Times(true)
	check(err)

	// host or machine kernel, uptime, platform Info
	hostStat, err := host.Info()
	check(err)

	// get interfaces MAC/hardware address
	//interfStat, err := net.Interfaces()
	//check(err)

	out, _ := json.Marshal(struct {
		Hostname  string `json:"hostname"`
		Timestamp int64  `json:"timestamp"`
		Etype     string `json:"etype"`
		Event     string `json:"event"`

		Host *host.InfoStat
		//Cpu   []cpu.InfoStat
		Times []cpu.TimesStat
		Vm    *mem.VirtualMemoryStat
		Usg   map[string]*disk.UsageStat
		//Net       []net.Interface
	}{hostname, now.Unix(), "health", "host", hostStat, times, v, usage})

	return out
}

// DefaultReporter is the front door for reporting
type DefaultReporter struct {
	// Errfn is the callback function that fires whenever there is an error.
	Errfn func(err error)
	gu.Producer

	hostname        string
	service         string
	version         string
	state           State
	message         string
	stats           sync.Map
	persistentStats sync.Map
	stdOutFallback  bool

	statFns             sync.Map
	healthCheckInterval time.Duration
	maxCheckInterval    time.Duration
	mux                 sync.Mutex
	reportTicker        *time.Ticker
	ctx                 context.Context
	cancel              context.CancelFunc
}

var _ Reporter = &DefaultReporter{}

// NewReporter create a new reporter
func NewReporter(service string, errfn func(err error), producer gu.Producer) *DefaultReporter {
	hn, _ := os.Hostname()
	info := build.GetInfo(service)

	ctx, cancel := context.WithCancel(context.Background())
	return &DefaultReporter{
		service:             service,
		hostname:            hn,
		version:             info.Version,
		state:               StateDefault,
		healthCheckInterval: time.Minute,
		maxCheckInterval:    MaxHealthCheckInterval,
		ctx:                 ctx,
		cancel:              cancel,
		reportTicker:        &time.Ticker{},
		Errfn:               errfn,
		Producer:            producer,
	}
}

// SetStdOutFallback defaults to false. When enabled, health messages will be written to std out if kafka is unhealthy
// or if no producer is set
func (r *DefaultReporter) SetStdOutFallback(b bool) {
	r.stdOutFallback = b
}

func (r *DefaultReporter) SetErrFn(efn func(err error)) {
	r.Errfn = efn
}

func (r *DefaultReporter) AddHostnameSuffix(suffix string) {
	if suffix == "" {
		return
	}

	r.hostname = fmt.Sprintf("%s-%s", r.hostname, suffix)
}

func (r *DefaultReporter) validate() error {
	if r.Producer == nil {
		return errors.New("producer cannot be nil")
	}
	return nil
}

func (r *DefaultReporter) Initialize() error {
	if err := r.validate(); err != nil {
		return err
	}

	r.ReportHealth()
	r.startHealthCheck(r.healthCheckInterval)
	return nil
}

// Health returns the current health as a populated Event object
func (r *DefaultReporter) Health() Event {
	r.mux.Lock()
	defer r.mux.Unlock()
	stats := make(map[string]interface{})
	fn := func(k interface{}, v interface{}) bool {
		if key, ok := k.(string); ok { // all keys should be strings, but just in case
			stats[key] = v
		} else {
			fmt.Printf("[warn] stat key of type (%T) is not string and will be ignored.\n", k)
		}
		return true
	}
	// persistent stats first, so we override them.
	r.persistentStats.Range(fn)
	r.stats.Range(fn)
	return Event{
		Hostname:  r.hostname,
		Timestamp: time.Now().Unix(),
		Type:      "health",
		Name:      "status",
		Service:   r.service,
		Version:   r.version,
		State:     r.state,
		Message:   r.message,
		Data:      stats,
	}
}

// SetHealth sets the current state and message
func (r *DefaultReporter) SetHealth(state State, message string) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.state = state
	r.message = message
}

// RegisterStatFn registers a function to be called before each call of ReportHealth.
func (r *DefaultReporter) RegisterStatFn(name string, fn func(reporter Reporter)) {
	r.statFns.Store(name, fn)
}

// ClearStatFns clears the reporter functions
func (r *DefaultReporter) ClearStatFns() {
	r.statFns = sync.Map{}
}

// PersistStat is akin to storing a stat, but a persistent stat is a stat that does not
// clear when Clear is called.
func (r *DefaultReporter) PersistStat(key string, val interface{}) {
	r.persistentStats.Store(key, val)
}

// AddStat adds a keyed entry to be included in reporting
func (r *DefaultReporter) AddStat(key string, val interface{}) {
	r.stats.Store(key, val)
}

// GetStat retrieve a previously entered stat or persistent stat.
// persistent stats will act as default stats.
func (r *DefaultReporter) GetStat(key string) interface{} {
	if val, ok := r.stats.Load(key); ok {
		return val
	}
	if val, ok := r.persistentStats.Load(key); ok {
		return val
	}
	return nil
}

// ClearStat clears an individual stat by key, does not clear persistent stats
func (r *DefaultReporter) ClearStat(key string) {
	r.stats.Delete(key)
}

// ClearStats clears all non persistent stats
func (r *DefaultReporter) ClearStats() {
	r.stats = sync.Map{}
}

// ClearAll clears all stats and persistent stats
func (r *DefaultReporter) ClearAll() {
	r.ClearStats()
	r.persistentStats = sync.Map{}
}

// JsonStats returns the current set of stats as a json
func (r *DefaultReporter) JsonStats() []byte {
	b := safeMarshal(r.Health().Data)
	return b
}

// ReportHealth sends the current health to kafka
func (r *DefaultReporter) ReportHealth() {
	r.statFns.Range(func(_, fun interface{}) bool {
		if fn, ok := fun.(func(reporter Reporter)); ok {
			fn(r)
		}
		return true
	})
	h := r.Health()
	b := safeMarshal(h)
	_ = r.produce(r.stdOutFallback, b) // we don't care about this error because it will get reported via callback
}

// StartIntervalReporting starts automatic reporting over the given interval.
func (r *DefaultReporter) StartIntervalReporting(interval time.Duration) {
	r.reportTicker.Stop() // just in case
	r.reportTicker = time.NewTicker(interval)
}

func (r *DefaultReporter) startHealthCheck(interval time.Duration) {
	healthTicker := time.NewTicker(interval)

	nextCheck := time.Time{}
	backOff := time.Duration(0)
	retry := 0
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				healthTicker.Stop()
				return
			case <-healthTicker.C:
				if time.Now().After(nextCheck) {
					if r.ProducerHealth() {
						retry = 0 // reset count when health
						backOff = time.Duration(0)
					} else {
						retry++
						backOff = backOff + (interval * time.Duration(retry)) // Super simple backoff logic,
						//log.Println("Backoff:", backOff)
						if backOff > r.maxCheckInterval {
							backOff = r.maxCheckInterval
						}
						nextCheck = time.Now().Add(backOff)
					}
				}
			case <-r.reportTicker.C:
				r.ReportHealth()
			}
		}
	}()
}

// Stop interval reporting and close the kafka producer, after calling stop, the health reporter should no
// longer be used. To re-start, you should create a new reporter
// calling this function will result in emitting a default shutdown state of Gray with an empty message
// if you want to set the final state or message, use StopWithFinalState
func (r *DefaultReporter) Stop() error {
	return r.StopWithFinalState(Gray, "")
}

// StopWithFinalState is the same as calling Stop() but takes in a final state and message,
// that will be emitted before shutdown.
func (r *DefaultReporter) StopWithFinalState(final State, msg string) error {
	r.SetHealth(final, msg)
	r.ReportHealth()
	if r.reportTicker != nil {
		r.reportTicker.Stop()
	}
	r.cancel()
	return nil
}

func (r *DefaultReporter) produce(fallback bool, message []byte) error {
	if err := r.Produce(message); err != nil {
		if fallback {
			// Fallback write to std out
			fmt.Println(string(message))
		}
		r.errorHandler("produce error", err)
		return err
	}
	return nil
}

// HealthHandler return the current health on demand
func (r *DefaultReporter) HealthHandler(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Query().Get("pretty")
	out := safeMarshal(r.Health())
	if pretty, _ := strconv.ParseBool(p); pretty {
		out = util.JsonPrettyPrint(out)
	}
	_, err := w.Write(out)
	// todo why call errorHandler here?
	r.errorHandler("could not write status response", err)
}

// todo it might be better to call this only when an error happens the first time,
// it's also possible we don't need this callback functionality
func (r *DefaultReporter) errorHandler(message string, err error) {
	go func() {
		if err != nil && r.Errfn != nil {
			r.Errfn(fmt.Errorf("%s: %v", message, err.Error()))
		}
	}()
}

func safeMarshal(msg interface{}) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		b, err = json.Marshal(struct {
			Error string `json:"error"` // *note* error doesn't marshal automatically
			Msg   string `json:"msg"`
		}{
			Error: "internal",
			Msg:   err.Error(),
		})
	}
	return b
}

func check(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}
