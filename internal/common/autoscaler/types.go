package autoscaler

type (
	// Resource is the unit of scalable resources
	Resource float64

	// Usages are different measurements used by a Recommender to provide a recommended Resource
	Usages map[UsageType]float64

	// UsageType type of usage
	UsageType string
)

const (
	// PollerUtilizationRate is a scale from 0 to 1 to indicate poller usages
	PollerUtilizationRate UsageType = "pollerUtilizationRate"
)

// Value helper method to return the value
func (r *Resource) Value() float64 {
	return float64(*r)
}
