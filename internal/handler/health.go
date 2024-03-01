package handler

type Healther interface {
	Health() error
}

type NamedHealther interface {
	Name() string
	Healther
}

type namedHealther struct {
	name string
	h    Healther
}

func (n *namedHealther) Name() string {
	return n.name
}

func (n *namedHealther) Health() error {
	return n.h.Health()
}

func NewNamedHealther(name string, h Healther) NamedHealther {
	return &namedHealther{name: name, h: h}
}

type HealthHandler struct {
	svc []NamedHealther
}

func NewHealthHandler(svc ...NamedHealther) *HealthHandler {
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
