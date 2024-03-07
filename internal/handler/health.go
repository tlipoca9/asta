package handler

import "github.com/tlipoca9/errors"

type Healthier interface {
	Health() error
}

type NamedHealthier interface {
	Name() string
	Healthier
}

type namedHealthier struct {
	name string
	h    Healthier
}

func (n *namedHealthier) Name() string {
	return n.name
}

func (n *namedHealthier) Health() error {
	if n.h == nil {
		return errors.New("healthier is nil")
	}
	return n.h.Health()
}

func NewNamedHealthier(name string, h Healthier) NamedHealthier {
	return &namedHealthier{name: name, h: h}
}

type HealthHandler struct {
	svc []NamedHealthier
}

func NewHealthHandler(svc ...NamedHealthier) *HealthHandler {
	return &HealthHandler{svc: svc}
}

func (h *HealthHandler) Health() (map[string]any, bool) {
	ret := make(map[string]any, len(h.svc))
	ok := true
	for _, s := range h.svc {
		if err := s.Health(); err != nil {
			ok = false
			ret[s.Name()] = map[string]any{
				"status": "down",
				"error":  err.Error(),
			}
		} else {
			ret[s.Name()] = map[string]any{
				"status": "up",
			}
		}
	}

	return ret, ok
}
