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

// Match returns true if the other service instance has the same attributes.
// StackID is ignored for correct comparing ServiceInstance got from database and ServiceInstance got from API request
func (i *ServiceInstance) Match(other *ServiceInstance) bool {
	return i.ID == other.ID &&
		i.ServiceID == other.ServiceID &&
		i.PlanID == other.PlanID &&
		reflect.DeepEqual(i.Params, other.Params)
}

// ServiceBinding represents a service binding.
type ServiceBinding struct {
	ID         string
	InstanceID string
	PolicyArn  string
	RoleName   string
	Scope      string
}

// Match returns true if the other service binding has the same attributes.
func (b *ServiceBinding) Match(other *ServiceBinding) bool {
	return b.ID == other.ID &&
		b.InstanceID == other.InstanceID &&
		b.RoleName == other.RoleName &&
		b.Scope == other.Scope
}
