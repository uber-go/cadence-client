package autoscaler

type (
	// Estimator collects data and estimate usage
	Estimator interface {
		CollectUsage(data interface{})
		Estimate() Usages
		Reset()
	}
)
