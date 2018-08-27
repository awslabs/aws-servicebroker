package serviceinstance

import "reflect"

// ServiceInstance provides details of a service instance
type ServiceInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]string
	StackID   string
}

func (i *ServiceInstance) Match(other *ServiceInstance) bool {
	return reflect.DeepEqual(i, other)
}
