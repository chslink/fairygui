package core

// Controller models a simple controller with page list.
type Controller struct {
	Name      string
	AutoRadio bool
	PageNames []string
	PageIDs   []string
}

// NewController constructs a controller.
func NewController(name string) *Controller {
	return &Controller{Name: name}
}
