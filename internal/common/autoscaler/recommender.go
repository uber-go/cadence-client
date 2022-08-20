package autoscaler

import "math"

// Recommender a recommendation generator for Resource
type Recommender interface {
	Recommend(currentResource Resource, currentUsages Usages) Resource
}

type linearRecommender struct {
	lower, upper  Resource
	targetMetrics Usages
}

// NewLinearRecommender create a linear Recommender
func NewLinearRecommender(lower, upper Resource, targetMetric Usages) Recommender {
	return &linearRecommender{
		lower:         lower,
		upper:         upper,
		targetMetrics: targetMetric,
	}
}

// Recommend recommends the new value
func (l *linearRecommender) Recommend(currentResource Resource, currentMetrics Usages) Resource {
	var res Resource
	for key := range currentMetrics {
		r := Resource(currentResource.Value() * currentMetrics[key] / l.targetMetrics[key])
		r = Resource(math.Min(l.upper.Value(), math.Max(l.lower.Value(), r.Value())))
		res += r
	}
	res /= Resource(len(currentMetrics))
	return res
}
