package work

import (
	"encoding/json"
	"expvar"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

type (
	h map[string]interface{}
	v []interface{}
)

func mockTime(sec int) func() time.Time {
	return func() time.Time {
		return time.Date(2017, 8, 11, 9, 0, sec, 0, time.UTC)
	}
}

func assertJSON(t *testing.T, o1, o2 interface{}) {
	var result, expect interface{}
	if reflect.TypeOf(o2).Kind() == reflect.Slice {
		result, expect = v{}, v{}
	} else {
		result, expect = h{}, h{}
	}
	if b1, err := json.Marshal(o1); err != nil {
		t.Fatal(o1, err)
	} else if err := json.Unmarshal(b1, &result); err != nil {
		t.Fatal(err)
	} else if b2, err := json.Marshal(o2); err != nil {
		t.Fatal(o2, err)
	} else if err := json.Unmarshal(b2, &expect); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(result, expect) {
		t.Fatal(result, expect)
	}
}

func TestCounter(t *testing.T) {
	c := NewMetricCounter()
	assertJSON(t, c, h{"type": "c", "count": 0})
	c.Add(1)
	assertJSON(t, c, h{"type": "c", "count": 1})
	c.Add(10)
	assertJSON(t, c, h{"type": "c", "count": 11})
}

func TestStatusGauge(t *testing.T) {
	sg := NewMetricStatusGauge(1, 2)
	assertJSON(t, sg, h{"type": "sg", "percentage": 0})
	sg.Add(500)
	assertJSON(t, sg, h{"type": "sg", "percentage": 1})
	sg.Add(200)
	assertJSON(t, sg, h{"type": "sg", "percentage": .5})
	sg.Add(200)
	assertJSON(t, sg, h{"type": "sg", "percentage": 0})
}

func TestMetricReset(t *testing.T) {
	c := &counter{}
	c.Add(5)
	assertJSON(t, c, h{"type": "c", "count": 5})
	c.Reset()
	assertJSON(t, c, h{"type": "c", "count": 0})

	sg := newStatusGauge(1, 2)
	assertJSON(t, sg, h{"type": "sg", "percentage": 0})
	sg.Add(500)
	assertJSON(t, sg, h{"type": "sg", "percentage": 1})
	sg.Reset()
	assertJSON(t, sg, h{"type": "sg", "percentage": 0})
}

func TestMetricString(t *testing.T) {
	c := NewMetricCounter()
	c.Add(1)
	c.Add(3)
	if s := c.String(); s != "4" {
		t.Fatal(s)
	}
	t.Log(c.String())

	sg := NewMetricStatusGauge(1, 2)
	sg.Add(500)
	sg.Add(200)
	if s := sg.String(); s != "0.5" {
		t.Fatal(s)
	}
}

func TestMetricValue(t *testing.T) {
	c := NewMetricCounter()
	c.Add(1)
	c.Add(3)
	if s := c.Value(); s != 4 {
		t.Fatal(s)
	}

	sg := NewMetricStatusGauge(1, 2)
	sg.Add(500)
	sg.Add(200)
	if s := sg.Value(); s != 0.5 {
		t.Fatal(s)
	}
}

func TestCounterTimeline(t *testing.T) {
	now = mockTime(0)
	c := NewMetricCounter("3s1s")
	expect := func(total float64, samples ...float64) h {
		timeline := v{}
		for _, s := range samples {
			timeline = append(timeline, h{"type": "c", "count": s})
		}
		return h{
			"interval": 1,
			"total":    h{"type": "c", "count": total},
			"samples":  timeline,
		}
	}
	assertJSON(t, c, expect(0, 0, 0, 0))
	c.Add(1)
	assertJSON(t, c, expect(1, 1, 0, 0))
	now = mockTime(1)
	assertJSON(t, c, expect(1, 0, 1, 0))
	c.Add(5)
	assertJSON(t, c, expect(6, 5, 1, 0))
	now = mockTime(3)
	assertJSON(t, c, expect(5, 0, 0, 5))
	now = mockTime(10)
	assertJSON(t, c, expect(0, 0, 0, 0))
}

func TestStatusGaugeTimeline(t *testing.T) {
	now = mockTime(0)
	sg := NewMetricStatusGauge(1, 2, "3s1s")
	expect := func(percentage float64, samples ...float64) h {
		timeline := v{}
		for _, s := range samples {
			timeline = append(timeline, h{"type": "sg", "percentage": s})
		}
		return h{
			"interval": 1,
			"total":    h{"type": "sg", "percentage": percentage},
			"samples":  timeline,
		}
	}

	assertJSON(t, sg, expect(0, 0, 0, 0))
	sg.Add(500)
	assertJSON(t, sg, expect(1, 1, 0, 0))
	now = mockTime(1)
	assertJSON(t, sg, expect(0, 0, 1, 0))
	sg.Add(200)
	sg.Add(200)
	assertJSON(t, sg, expect(0, 0, 1, 0))
	now = mockTime(3)
	assertJSON(t, sg, expect(0, 0, 0, 0))
	now = mockTime(10)
	j, _ := json.Marshal(sg)
	t.Log(string(j))
	assertJSON(t, sg, expect(0, 0, 0, 0))
}

func TestMulti(t *testing.T) {
	m := NewMetricCounter("10s1s", "30s5s")
	m.Add(5)
	if s := m.String(); s != `5` {
		t.Fatal(s)
	}
	if s := m.Value(); s != 5 {
		t.Fatal(s)
	}

	sg := NewMetricStatusGauge(1, 2, "10s1s", "30s5s")
	sg.Add(500)
	sg.Add(200)
	if s := sg.Value(); s != 0.5 {
		t.Fatal(s)
	}
}

func TestExpVar(t *testing.T) {
	expvar.Publish("test:count", NewMetricCounter())
	expvar.Get("test:count").(Metric).Add(1)
	if s := expvar.Get("test:count").String(); s != `1` {
		t.Fatal(s)
	}

	expvar.Publish("test:statusgauge", NewMetricStatusGauge(1, 2))
	expvar.Get("test:statusgauge").(Metric).Add(500)
	expvar.Get("test:statusgauge").(Metric).Add(200)
	if s := expvar.Get("test:statusgauge").String(); s != `0.5` {
		t.Fatal(s)
	}
}

func BenchmarkMetrics(b *testing.B) {
	b.Run("counter", func(b *testing.B) {
		c := &counter{}
		for i := 0; i < b.N; i++ {
			c.Add(rand.Float64())
		}
	})
	b.Run("timeline/counter", func(b *testing.B) {
		c := NewMetricCounter("10s1s")
		for i := 0; i < b.N; i++ {
			c.Add(rand.Float64())
		}
	})
}
