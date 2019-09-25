package work

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func Test_WithHealthCheck(t *testing.T) {
	is := is.New(t)

	logger := logrus.WithFields(logrus.Fields{
		"method": "TestInitDefaultTopicPublisher",
	})
	r := gin.Default()

	health := NewHealthCheck(WithEngine(r), WithHealthTicker(time.NewTicker(2*time.Millisecond)))
	defer health.Close()

	healthAdapter := WithHealthCheck(health, WithErrorPercentageByRequestCount(float64(.25), 3, 4))
	badhandler := Adapt(func(j *Job) error {
		j.Ctx.SetStatus(StatusInternalError)
		return nil
	}, healthAdapter)

	goodhandler := Adapt(func(j *Job) error {
		j.Ctx.SetStatus(StatusSuccess)
		return nil
	}, healthAdapter)

	w := NewCommonWorker(logger)
	err := w.Register(
		"bad",
		badhandler)
	is.NoErr(err)
	err = w.Register(
		"good",
		goodhandler)
	is.NoErr(err)
	err = w.Perform(
		&Job{Handler: "bad"},
		WithSync(true))
	is.NoErr(err)
	is.True(health.GetStatus() == 200)
	for i := 0; i < 10; i++ {
		err = w.Perform(
			&Job{Handler: "bad"},
			WithSync(true))
		is.NoErr(err)
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(health.GetStatus())
	is.True(health.GetStatus() == 500)

	for i := 0; i < 3; i++ {
		err = w.Perform(
			&Job{Handler: "good"},
			WithSync(true))
		is.NoErr(err)
		time.Sleep(20 * time.Millisecond)
	}
	t.Log(health.GetStatus())
	is.True(health.GetStatus() == 200)

}
