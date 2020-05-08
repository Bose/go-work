package work

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WithHealthCheck is an adpater middleware for healthcheck.
// it also adds a health http GET endpoint.  It supports the Option.
// WithErrorPercentage(percentageOfErrors float64, lastNumOfMinutes int) that allows
// you to override the default of: 1.0 (100%) errors in the last 5 min.
func WithHealthCheck(health *HealthCheck, opt ...Option) Adapter {
	health.Engine.GET(health.HealthPath, health.Handler)
	health.SetStatus(http.StatusOK) // set status to OK as we get started

	opts := GetOpts(opt...)

	// setup some defaults for the options
	useMinutes := true
	lastNumRequests := 15
	lastNumMinutes := 15
	minNumOfRequests := 5
	allowedErrPercentage := float64(1.0)

	if v, ok := opts[optionMinNumOfRequests].(int); ok {
		minNumOfRequests = v
	}
	if v, ok := opts[optionLastNumOfRequests].(int); ok {
		lastNumRequests = v
		useMinutes = false
	}
	if p, ok := opts[optionPercentageOfErrors].(float64); ok {
		allowedErrPercentage = p
	}
	if m, ok := opts[optionLastNumOfMinutes].(int); ok {
		lastNumMinutes = m
	}
	frame := fmt.Sprintf("%dm%dm", lastNumMinutes, lastNumMinutes) // just define one frame (time window)
	health.requestCnt = NewMetricCounter(frame)
	health.errorCnt = NewMetricCounter(frame)

	health.numReqGauge = NewMetricStatusGauge(minNumOfRequests, lastNumRequests)

	var totalRequestsMU sync.Mutex
	totalRequests := 0
	return func(h Handler) Handler {
		return Handler(func(job *Job) error {
			totalRequestsMU.Lock()
			currentTotalRequests := totalRequests
			totalRequestsMU.Unlock()

			if currentTotalRequests <= minNumOfRequests {
				totalRequestsMU.Lock()
				totalRequests += 1
				currentTotalRequests = totalRequests
				totalRequestsMU.Unlock()
			}
			if useMinutes {
				health.requestCnt.Add(1)
			}
			// do health stuff...
			defer func() {
				status := job.Ctx.Status()
				if useMinutes {
					if isErrorStatus(status) {
						health.errorCnt.Add(1)
					}
				} else {
					health.numReqGauge.Add(float64(status))
				}
				select {
				case <-health.metricTicker.C:
					if currentTotalRequests > minNumOfRequests { // make sure we've seen the min number of requests before starting to calc
						var badPercentage float64
						if useMinutes {
							// you can just get the Value() of the time series metric since there's only one frame.... multi frames won't work this way
							badPercentage = health.errorCnt.Value() / health.requestCnt.Value()
						} else {
							badPercentage = health.numReqGauge.Value()
						}
						if badPercentage > allowedErrPercentage {
							end := time.Now()
							end.UTC()
							logger := logrus.WithFields(logrus.Fields{
								"method": "WithHealthCheck",
								"time":   end.Format(time.RFC3339),
							})
							logger.Errorf("WithHealthCheck: unhealthy error percentage of %f and only allowed %f", badPercentage, allowedErrPercentage)
							health.SetStatus(http.StatusInternalServerError)
						} else {
							health.SetStatus(http.StatusOK)
						}

					}
				default:
				}
			}()
			err := h(job)
			return err
		})
	}
}

func isErrorStatus(status Status) bool {
	if status == StatusTimeout || status == StatusInternalError || status == StatusUnavailable {
		return true
	}
	return false
}

// healthStatus is a goroutine safe way to set/get the health status
type healthStatus struct {
	sync.RWMutex
	status int
}

func (h *healthStatus) setStatus(s int) {
	h.Lock()
	defer h.Unlock()
	h.status = s
}
func (h *healthStatus) getStatus() int {
	h.Lock()
	defer h.Unlock()
	return h.status
}

// HealthCheck provides a healthcheck endpoint for the work
type HealthCheck struct {

	// HealthPath is the GET url path for the endpoint
	HealthPath string

	// Engine is the gin.Engine that should serve the endpoint
	Engine *gin.Engine

	// Handler is the gin Hanlder to use for the endpoint
	Handler gin.HandlerFunc

	// status is the result to return when the endpoint is called
	status *healthStatus

	metricTicker *time.Ticker
	requestCnt   Metric
	errorCnt     Metric

	numReqGauge Metric
}

// NewHealthCheck creates a new HealthCheck with the options provided.
// Options: WithEngine(*gin.Engine), WithHealthPath(string), WithHealthHander(gin.HandlerFunc), WithMetricTicker(time.Ticker)
func NewHealthCheck(opt ...Option) *HealthCheck {
	opts := GetOpts(opt...)

	h := &HealthCheck{
		status: &healthStatus{},
	}
	if e, ok := opts[optionWithEngine].(*gin.Engine); ok {
		h.Engine = e
	}
	if path, ok := opts[optionWithHealthPath].(string); ok {
		h.HealthPath = path
	} else {
		h.HealthPath = "/ready"
	}
	if handler, ok := opts[optionWithHealthHandler].(gin.HandlerFunc); ok {
		h.Handler = handler
	} else {
		h.Handler = h.DefaultHealthHandler()
	}

	if ticker, ok := opts[optionHealthTicker].(*time.Ticker); ok {
		h.metricTicker = ticker
	} else {
		h.metricTicker = time.NewTicker(DefaultHealthTickerDuration)
	}

	return h
}

// Close cleans up the all the HealthCheck resources
func (h *HealthCheck) Close() {
	h.metricTicker.Stop()
}

// GetStatus returns the current health status
func (h *HealthCheck) GetStatus() int {
	return h.status.getStatus()
}

// SetStatus sets the current health status
func (h *HealthCheck) SetStatus(s int) {
	h.status.setStatus(s)
}

// WithEngine lets you set the *gin.Engine if it's created after you've created the *HealthCheck
func (h *HealthCheck) WithEngine(e *gin.Engine) {
	h.Engine = e
}

func (h *HealthCheck) DefaultHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		status := h.GetStatus()
		if status == 500 {
			end := time.Now()
			end.UTC()
			logger := logrus.WithFields(logrus.Fields{
				"method": "WithHealthCheck",
				"time":   end.Format(time.RFC3339),
			})
			logger.Errorf("Reporting Health: DOWN")
			c.String(h.GetStatus(), "DOWN")
		} else {
			c.String(h.GetStatus(), "OK")
		}
		return
	}
}

// DefaultHealthTickerDuration is the time duration between the recalculation of the status returned by HealthCheck.GetStatus()
const DefaultHealthTickerDuration = 1 * time.Minute
const optionHealthTicker = "optionHealthTicker"

func WithHealthTicker(ticker *time.Ticker) Option {
	return func(o Options) {
		o[optionHealthTicker] = ticker
	}
}

const optionWithHealthHandler = "optionWithHealthHandler"

// WithHealthHandler override the default health endpoint handler
func WithHealthHandler(h gin.HandlerFunc) Option {
	return func(o Options) {
		o[optionWithHealthHandler] = h
	}
}

const optionWithHealthPath = "optionWithHealthPath"

// WithHealthPath override the default path for the health endpoint
func WithHealthPath(path string) Option {
	return func(o Options) {
		o[optionWithHealthPath] = path
	}
}

const optionPercentageOfErrors = "optionPercentageOfErrors"
const optionLastNumOfMinutes = "optionLastNumOfMinutes"

// WithErrorPercentage allows you to override the default of 1.0 (100%) with the % you want for error rate.
func WithErrorPercentage(percentageOfErrors float64, lastNumOfMinutes int) Option {
	return func(o Options) {
		o[optionPercentageOfErrors] = percentageOfErrors
		o[optionLastNumOfMinutes] = lastNumOfMinutes
	}

}

const optionLastNumOfRequests = "optionLastNumOfRequests"
const optionMinNumOfRequests = "optionMinNumOfRequests"

func WithErrorPercentageByRequestCount(percentageOfErrors float64, minNumOfRequests, lastNumOfRequests int) Option {
	return func(o Options) {
		o[optionPercentageOfErrors] = percentageOfErrors
		o[optionLastNumOfRequests] = lastNumOfRequests
		o[optionMinNumOfRequests] = minNumOfRequests
	}
}
