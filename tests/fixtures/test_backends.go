package fixtures

import "github.com/reybrally/TCP-Load-Balancer/internal/domain/model"

func CreateTestBackends(count int) []*model.Backend {
	backends := make([]*model.Backend, count)
	for i := 0; i < count; i++ {
		backends[i] = model.NewBackend(
			"backend-"+string(rune('1'+i)),
			"localhost",
			3001+i,
			1,
		)
	}
	return backends
}
