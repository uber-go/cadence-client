package autoscaler

// AutoScaler collects data and estimate usage
type (
	AutoScaler interface {
		Estimator
		// Acquire X unit of Resource
		Acquire(Resource) error
		// Release X unit of Resource
		Release(Resource)
		// Start starts the autoscaler go routine that
		Start() DoneFunc
	}
	// DoneFunc func to turn off auto scaler
	DoneFunc func()
)
