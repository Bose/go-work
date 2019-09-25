package work

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func Test_RunE(t *testing.T) {
	logger := logrus.WithFields(logrus.Fields{
		"method": "Test_WithGinLogrus",
	})
	is := is.New(t)
	ctx := NewContext(context.Background())
	ctx.Set("test-key", []string{"hi", "mom"})
	job := &Job{
		Handler: "test-handler",
		Args: Args{
			"foo": "bar",
		},
		Ctx: &ctx,
	}
	logger.Info("testing WithJob and expect panic")
	err := RunE(func() error {
		panic("test panic")
	},
		WithJob(job),
	)
	t.Log(job.Ctx.Status())
	is.True(job.Ctx.Status() == StatusInternalError)
	is.True(err != nil)
	t.Log(err)

	logger.Info("testing with NO job and expect panic")
	err = RunE(func() error {
		panic("test panic")
	})
	is.True(err != nil)
	t.Log(err)

	logger.Info("expecting no panic")
	err = RunE(func() error {
		t.Log("success")
		return nil
	})
	is.NoErr(err)
}
