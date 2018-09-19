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
