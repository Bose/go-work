package work

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// To mock time in tests
var now = time.Now

// Metric is a single meter (a counter for now, but in the future: gauge or histogram, optionally - with history)
type Metric interface {
	Add(n float64)
	String() string
	Value() float64
}

// metric is an extended private interface with some additional internal
// methods used by timeseries. Counters implement it for now, but in the future: gauges and histograms
type metric interface {
	Metric
	Reset()
	Aggregate(roll int, samples []metric)
}

var _ metric = &counter{}

// NewCounter returns a counter metric that increments the value with each
// incoming number.
func NewMetricCounter(frames ...string) Metric {
	return newMetric(func() metric { return &counter{} }, frames...)
}

type timeseries struct {
	sync.Mutex
	now      time.Time
	interval time.Duration
	total    metric
	samples  []metric
}

// Reset the timeseries
func (ts *timeseries) Reset() {
	ts.total.Reset()
	for _, s := range ts.samples {
		s.Reset()
	}
}

func (ts *timeseries) roll() {
	t := now()
	roll := int((t.Round(ts.interval).Sub(ts.now.Round(ts.interval))) / ts.interval)
	ts.now = t
	n := len(ts.samples)
	if roll <= 0 {
		return
	}
	if roll >= len(ts.samples) {
		ts.Reset()
	} else {
		for i := 0; i < roll; i++ {
			tmp := ts.samples[n-1]
			for j := n - 1; j > 0; j-- {
				ts.samples[j] = ts.samples[j-1]
			}
			ts.samples[0] = tmp
			ts.samples[0].Reset()
		}
		ts.total.Aggregate(roll, ts.samples)
	}
}

// Add to the timeseries
func (ts *timeseries) Add(n float64) {
	ts.Lock()
	defer ts.Unlock()
	ts.roll()
	ts.total.Add(n)
	ts.samples[0].Add(n)
}

// MarshalJSON from the timeseries
func (ts *timeseries) MarshalJSON() ([]byte, error) {
	ts.Lock()
	defer ts.Unlock()
	ts.roll()
	return json.Marshal(struct {
		Interval float64  `json:"interval"`
		Total    Metric   `json:"total"`
		Samples  []metric `json:"samples"`
	}{float64(ts.interval) / float64(time.Second), ts.total, ts.samples})
}

// String representation of the timeseries
func (ts *timeseries) String() string {
	ts.Lock()
	defer ts.Unlock()
	ts.roll()
	return ts.total.String()
}

// Value of the timeseries
func (ts *timeseries) Value() float64 {
	ts.Lock()
	defer ts.Unlock()
	ts.roll()
	return ts.total.Value()
}

type multimetric []*timeseries

// Add to the multimetric
func (mm multimetric) Add(n float64) {
	for _, m := range mm {
		m.Add(n)
	}
}

// MarshalJSON from the multimetric
func (mm multimetric) MarshalJSON() ([]byte, error) {
	b := []byte(`{"metrics":[`)
	for i, m := range mm {
		if i != 0 {
			b = append(b, ',')
		}
		x, _ := json.Marshal(m)
		b = append(b, x...)
	}
	b = append(b, ']', '}')
	return b, nil
}

// String representation of the multimetric
func (mm multimetric) String() string {
	return mm[len(mm)-1].String()
}

// Value of the multimetric
func (mm multimetric) Value() float64 {
	return mm[len(mm)-1].Value()
}

type counter struct {
	count uint64
}

// String representation of a counter
func (c *counter) String() string {
	return strconv.FormatFloat(c.Value(), 'g', -1, 64)
}

// Reset the counter
func (c *counter) Reset() {
	atomic.StoreUint64(&c.count, math.Float64bits(0))
}

// Value of the counter
func (c *counter) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.count))
}

// Add to the couter
func (c *counter) Add(n float64) {
	for {
		old := math.Float64frombits(atomic.LoadUint64(&c.count))
		new := old + n
		if atomic.CompareAndSwapUint64(&c.count, math.Float64bits(old), math.Float64bits(new)) {
			return
		}
	}
}

// MarshalJSON from the counter
func (c *counter) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string  `json:"type"`
		Count float64 `json:"count"`
	}{"c", c.Value()})
}

// Aggregate the counter
func (c *counter) Aggregate(roll int, samples []metric) {
	c.Reset()
	for _, s := range samples {
		c.Add(s.(*counter).Value())
	}
}

// make sure statusGauge conforms to the Metric interface
var _ metric = &statusGauge{}

type statusGauge struct {
	statuses []Status
	min      int
	max      int
	moot     *sync.Mutex
}

// NewMetricStatusGauge is a factory for statusGauge Metrics
func NewMetricStatusGauge(min int, max int, frames ...string) Metric {
	return newMetric(func() metric { return newStatusGauge(min, max) }, frames...)
}

// newStatusGauge factory for *statusGauge
func newStatusGauge(min int, max int) *statusGauge {
	if min <= 0 {
		min = 1
	}
	if max < min {
		max = min
	}
	// warning: do not try and do something like: statuses: make([]Status, max)
	// this will create issues since len(statuses) has meaning in this implementation
	s := statusGauge{
		max:  max,
		min:  min,
		moot: &sync.Mutex{},
	}
	return &s
}

// Aggregate the statusGauge
func (g *statusGauge) Aggregate(roll int, samples []metric) {
	g.Reset()
	for _, s := range samples {
		g.Add(s.(*statusGauge).Value())
	}
}

// String repr of the statusGauge
func (g *statusGauge) String() string {
	return strconv.FormatFloat(g.Value(), 'g', -1, 64)
}

// Value is the percentage of the statusGauge
func (g *statusGauge) Value() float64 {
	return g.percentage()
}

// Add pushes a new value into the statusGauge
func (g *statusGauge) Add(f float64) {
	g.push(Status(f))
}

// Reset the statusGauge
func (g *statusGauge) Reset() {
	g.statuses = g.statuses[:0] // just truncate the slice
}

// MarshalJSON from statusGauge
func (g *statusGauge) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type       string  `json:"type"`
		Percentage float64 `json:"percentage"`
	}{"sg", g.Value()})
}

// percentage of errors in the statusGauge
func (g *statusGauge) percentage() float64 {
	g.moot.Lock()
	defer g.moot.Unlock()
	if len(g.statuses) < g.min {
		return float64(0)
	}
	errCnt := 0
	for _, s := range g.statuses {
		if isErrorStatus(s) {
			errCnt += 1
		}
	}
	return float64(errCnt) / float64(len(g.statuses))
}

// push a new Status into the statusGauge
func (g *statusGauge) push(s Status) {
	if g.max == g.len() {
		g.pop()
	}
	g.moot.Lock()
	g.statuses = append(g.statuses, s)
	g.moot.Unlock()
}

func (g *statusGauge) len() int {
	g.moot.Lock()
	defer g.moot.Unlock()
	return len(g.statuses)
}

// pop a Status off the statusGauge
func (g *statusGauge) pop() Status {
	g.moot.Lock()
	defer g.moot.Unlock()
	v := g.statuses[0]
	g.statuses = g.statuses[1:]
	return v
}

func newTimeseries(builder func() metric, frame string) *timeseries {
	var (
		totalNum, intervalNum   int
		totalUnit, intervalUnit rune
	)
	units := map[rune]time.Duration{
		's': time.Second,
		'm': time.Minute,
		'h': time.Hour,
		'd': time.Hour * 24,
		'w': time.Hour * 24 * 7,
		'M': time.Hour * 24 * 30,
		'y': time.Hour * 24 * 365,
	}
	fmt.Sscanf(frame, "%d%c%d%c", &totalNum, &totalUnit, &intervalNum, &intervalUnit)
	interval := units[intervalUnit] * time.Duration(intervalNum)
	if interval == 0 {
		interval = time.Minute
	}
	totalDuration := units[totalUnit] * time.Duration(totalNum)
	if totalDuration == 0 {
		totalDuration = interval * 15
	}
	n := int(totalDuration / interval)
	samples := make([]metric, n, n)
	for i := 0; i < n; i++ {
		samples[i] = builder()
	}
	totalMetric := builder()
	return &timeseries{interval: interval, total: totalMetric, samples: samples}
}

func newMetric(builder func() metric, frames ...string) Metric {
	if len(frames) == 0 {
		return builder()
	}
	if len(frames) == 1 {
		return newTimeseries(builder, frames[0])
	}
	mm := multimetric{}
	for _, frame := range frames {
		mm = append(mm, newTimeseries(builder, frame))
	}
	sort.Slice(mm, func(i, j int) bool {
		a, b := mm[i], mm[j]
		return a.interval.Seconds()*float64(len(a.samples)) < b.interval.Seconds()*float64(len(b.samples))
	})
	return mm
}
