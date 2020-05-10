package health

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Setheck/gu/health/mocks"
	"github.com/Setheck/gu/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaults(t *testing.T) {

	assert.Equal(t, "green", string(Green))
	assert.Equal(t, "blue", string(Blue))
	assert.Equal(t, "yellow", string(Yellow))
	assert.Equal(t, "red", string(Red))
	assert.Equal(t, "gray", string(Gray))

}

func generateFields(msg Event) (map[string]interface{}, error) {
	output, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	err = json.Unmarshal(output, &fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

var requiredFields = [...]string{"hostname", "timestamp", "etype", "event", "service", "version", "state", "msg"}

func TestRequiredMessageFields(t *testing.T) {

	fields, err := generateFields(Event{})
	assert.Equal(t, err, nil)

	for _, field := range requiredFields {
		_, ok := fields[field]

		assert.Equal(t, ok, true)
	}

}

var reservedFields = [...]string{"data"}

func TestReservedMessageFields(t *testing.T) {
	fields, err := generateFields(Event{})
	assert.Equal(t, err, nil)

	for _, field := range reservedFields {
		_, ok := fields[field]

		assert.Equal(t, ok, false)
	}
}

func TestOptionalDataField(t *testing.T) {

	data := make(map[string]interface{})

	data["magic"] = Green

	fields, err := generateFields(Event{Data: data})
	assert.Equal(t, err, nil)

	data = fields["data"].(map[string]interface{})

	assert.Equal(t, data["magic"].(string), string(Green))
}

func TestNewReporter(t *testing.T) {
	mockProducer := new(mocks.Producer)
	r := NewReporter("service", func(err error) {}, mockProducer)
	m := r.Health()
	h, _ := os.Hostname()
	assert.Equal(t, h, m.Hostname)
	assert.NotZero(t, m.Timestamp)
	assert.Equal(t, "health", m.Type)
	assert.Equal(t, "status", m.Name)
	assert.Equal(t, "service", m.Service)
	assert.NotEmpty(t, m.Version)
	assert.Equal(t, StateDefault, m.State)
	assert.Equal(t, "", m.Message)
	assert.IsType(t, map[string]interface{}{}, m.Data)
	assert.NoError(t, r.Stop())
}

func TestSetHealth(t *testing.T) {
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(err error) {}, mockProducer)
	assert.Equal(t, Blue, r.Health().State)
	r.SetHealth(Green, "go green!")

	assert.Equal(t, Green, r.Health().State)
	assert.Equal(t, "go green!", r.Health().Message)

	assert.NoError(t, r.Stop())
	assert.Equal(t, Gray, r.Health().State)
}

func TestHealthStats(t *testing.T) {
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(err error) {}, mockProducer)

	// set a stat
	r.AddStat("key", "val")
	data, ok := r.Health().Data.(map[string]interface{})
	if !ok {
		assert.Fail(t, "invalid Data field type")
	}
	assert.Equal(t, "val", data["key"])

	// persist the same stat, but since it already has a value. It should just get masked
	r.PersistStat("key", "default")
	data, ok = r.Health().Data.(map[string]interface{})
	if !ok {
		assert.Fail(t, "invalid Data field type")
	}
	assert.Equal(t, "val", data["key"]) // persistent stat is always lower in priority to regular stat

	r.ClearStats() // clears regular stats

	data, ok = r.Health().Data.(map[string]interface{})
	if !ok {
		assert.Fail(t, "invalid Data field type")
	}
	assert.Equal(t, "default", data["key"]) // verify persistent

	assert.NoError(t, r.Stop())
}

func TestHostnameSuffix(t *testing.T) {
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(err error) {}, mockProducer)

	suffix := "test.suffix"

	hn := r.Health().Hostname
	r.AddHostnameSuffix(suffix)
	nhn := r.Health().Hostname

	assert.True(t, strings.HasSuffix(nhn, fmt.Sprintf("-%s", suffix)))

	assert.Equal(t, nhn, fmt.Sprintf("%s-%s", hn, suffix))

	assert.NoError(t, r.Stop())
}

func TestStats(t *testing.T) {
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(err error) {}, mockProducer)

	h := r.Health()
	assert.Empty(t, h.Data)

	r.AddStat("one", "s")
	r.AddStat("two", "s")
	assert.Equal(t, "s", r.GetStat("one"))

	r.ClearStat("one")
	assert.Nil(t, r.GetStat("one"))
	assert.NotNil(t, r.GetStat("two"))

	r.ClearStats()
	assert.Nil(t, r.GetStat("two"))

	assert.NoError(t, r.Stop())
}

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		assert.Error(t, err)
	}

	rr := httptest.NewRecorder()

	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(err error) {}, mockProducer)

	handler := http.HandlerFunc(r.HealthHandler)

	// keeping these two calls as close as possible to prevent timestamp changes
	want := r.Health()
	handler.ServeHTTP(rr, req)

	var got Event
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, want, got)

	assert.NoError(t, r.Stop())
}

func TestHealthHandlerPretty(t *testing.T) {
	req, err := http.NewRequest("GET", "/?pretty=true", nil)
	if err != nil {
		assert.Error(t, err)
	}

	rr := httptest.NewRecorder()

	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(err error) {}, mockProducer)

	handler := http.HandlerFunc(r.HealthHandler)

	// keeping these two calls as close as possible to prevent timestamp changes
	h := r.Health()
	handler.ServeHTTP(rr, req)

	// Marshal and prettify
	b, err := json.Marshal(h)
	if err != nil {
		assert.Error(t, err)
	}
	want := util.JsonPrettyPrint(b)

	// got should already be marshaled and pretty because 'pretty=true' query param
	got := rr.Body.Bytes()

	assert.Equal(t, want, got)

	assert.NoError(t, r.Stop())
}

func TestReporter_Stop(t *testing.T) {
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", func(error) {}, mockProducer)
	mockProducer.On("Produce", mock.AnythingOfType("*string"), mock.MatchedBy(func(b []uint8) bool {
		evt := unmarshalEvent(t, b)

		// verify message and state match the default state
		return evt.Message == "" && evt.State == Gray
	})).Return(nil)

	assert.NoError(t, r.Stop())
}

func TestReporter_StopWithFinalState(t *testing.T) {
	tests := []struct {
		state State
		msg   string
	}{
		{Blue, util.RandomString(10)},
		{Green, util.RandomString(10)},
		{Yellow, util.RandomString(10)},
		{Red, util.RandomString(10)},
		{Gray, util.RandomString(10)},
	}
	for _, test := range tests {
		t.Run(string(test.state), func(t *testing.T) {
			mockProducer := new(mocks.Producer)
			r := NewReporter("test", func(error) {}, mockProducer)
			mockProducer.On("Produce", mock.MatchedBy(func(b []uint8) bool {
				evt := unmarshalEvent(t, b)
				// verify message and state get matched in the produced event
				return evt.Message == test.msg && evt.State == test.state
			})).Return(nil)

			assert.NoError(t, r.StopWithFinalState(test.state, test.msg))
		})
	}
}

func TestPersistentStats(t *testing.T) {
	errFn := func(error) {}
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", errFn, mockProducer)
	// *note* no need to initialize because we aren't touching kafka

	assert.Empty(t, r.GetStat("key"))

	r.PersistStat("key", "default")

	assert.Equal(t, "default", r.GetStat("key"))

	r.AddStat("key", "value")

	assert.Equal(t, "value", r.GetStat("key"))

	r.ClearStat("key")

	assert.Equal(t, "default", r.GetStat("key"))

	r.AddStat("key", "value2")

	r.ClearStats()

	assert.Equal(t, "default", r.GetStat("key"))

	r.ClearAll()

	assert.Empty(t, r.GetStat("key"))

	assert.NoError(t, r.Stop())
}

func TestBackOff(t *testing.T) {
	t.SkipNow()
	errFn := func(err error) { fmt.Println("err:", err) }
	mockProducer := new(mocks.Producer)
	r := NewReporter("test", errFn, mockProducer)
	r.maxCheckInterval = time.Second * 5
	r.healthCheckInterval = time.Millisecond * 100
	err := r.Initialize()
	assert.Error(t, err)

	//r.SetStdOutFallback(true)
	for i := 0; i < 2000; i++ {
		//fmt.Println("Calling Write")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Time:%s\n", time.Now())
		//time.Sleep(time.Second)
	}

	assert.NoError(t, r.Stop())
}

func TestReporter_AddFunc(t *testing.T) {
	mockProducer := new(mocks.Producer)
	errFn := func(err error) { fmt.Println("err:", err) }
	r := NewReporter("test", errFn, mockProducer)

	mockProducer.On("Produce", mock.AnythingOfType("[]uint8")).Return(nil)

	key := util.RandomString(10)
	val := util.RandomString(10)
	r.RegisterStatFn("test", func(reporter Reporter) {
		r.PersistStat(key, val)
	})
	r.ReportHealth()
	mockProducer.AssertCalled(t, "Produce", mock.MatchedBy(func(b []byte) bool {
		var evt Event
		if err := json.Unmarshal(b, &evt); err != nil {
			t.Fatal(err)
		}
		if data, ok := evt.Data.(map[string]interface{}); ok {
			return data[key] == val
		}
		return false
	}))
}

func TestRunReporter(t *testing.T) {
	t.SkipNow()
	/*
		This is example usage of the health reporter
		as well as a test that can be pointed to kafka
		and run for a simple manual verification.
	*/

	// Define an error function for notification when an error occurs
	// *Note* since we have `log.SetOutput(r.GetKafkaLogWriter())` below,
	// we cannot use `log.Println` in the error handler.
	eFn := func(err error) {
		fmt.Println("Error Occurred:", err)
	}

	mockProducer := new(mocks.Producer)
	// Create a new health reporter
	r := NewReporter("test", eFn, mockProducer)

	// Initialize is called to create a new producer.
	if err := r.Initialize(); err != nil {
		t.Error(err)
	}

	// Start interval reporting every second
	r.StartIntervalReporting(time.Second)
	time.Sleep(time.Second)

	// Clear the old stat, and add a new one.
	r.PersistStat("Stat", "a default value")
	time.Sleep(time.Second)

	// Add stats that should show up on the next report
	r.AddStat("Stat", "Woo")
	time.Sleep(time.Second)

	// clearing this stat results back to the default value
	r.ClearStat("Stat")
	r.AddStat("new", "stat")
	time.Sleep(time.Second)

	// clear all but persistent stats
	r.ClearStats()
	time.Sleep(time.Second)

	r.ClearAll()
	time.Sleep(time.Second)

	// If brokers are down, we need to have StdOutFallback enabled to redirect the logWriter to std.Out
	log.Println("testing with fallback!")
	r.SetStdOutFallback(true)
	r.ReportHealth()
	time.Sleep(time.Second)

	time.Sleep(time.Second)
	// Be sure to stop the the health reporter as a shutdown step.
	assert.NoError(t, r.Stop())
}

func TestSafeMarshal(t *testing.T) {
	goodOut := safeMarshal(struct {
		One   string
		Two   int
		Three float32
	}{
		"one",
		2,
		3.0,
	})
	assert.Equal(t, `{"One":"one","Two":2,"Three":3}`, string(goodOut))

	badOutput := safeMarshal(struct {
		Chan chan int // can't serialize a channel
	}{
		Chan: make(chan int),
	})
	assert.Equal(t, `{"error":"internal","msg":"json: unsupported type: chan int"}`, string(badOutput))
}

func unmarshalEvent(t *testing.T, b []byte) Event {
	t.Helper()
	var evt Event
	err := json.Unmarshal(b, &evt)
	if err != nil {
		t.Error(err)
	}
	return evt
}

func TestA(t *testing.T) {
	errFn := func(err error) { fmt.Println("Reporter Error:", err) }
	producerMock := new(mocks.Producer)
	hr := NewReporter("test", errFn, producerMock)

	if err := hr.Initialize(); err != nil {
		t.Fatal(err)
	}

	hr.StartIntervalReporting(time.Second)

	<-time.After(time.Minute * 5)
	if err := hr.Stop(); err != nil {
		t.Fatal(err)
	}
}
