package work

// Status for a job's execution (Perform)
type Status int

const (
	// StatusUnknown was a job with an unknown status
	StatusUnknown Status = -1

	// StatusSucces was a successful job
	StatusSuccess = 200

	// StatusBadRequest was a job with a bad request
	StatusBadRequest = 400

	// StatusForbidden was a forbidden job
	StatusForbidden = 403

	// StatusUnauthorized was an unauthorized job
	StatusUnauthorized = 401

	// StatusTimeout was a job that timed out
	StatusTimeout = 408

	// StatusNoResponse was a job that intentionally created no response (basically the conditions were met for a noop by the Job)
	StatusNoResponse = 444

	// StatusInternalError was a job with an internal error
	StatusInternalError = 500

	// StatusUnavailable was a job that was unavailable
	StatusUnavailable = 503
)
