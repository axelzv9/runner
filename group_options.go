package runner

type GroupOptions func(*group)

func WithPool(pool *Pool) GroupOptions {
	return func(g *group) {
		g.pool = pool
	}
}
