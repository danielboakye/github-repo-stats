package httpserver

import "github.com/go-chi/chi"

// RegisterRoutes setups routes for http server
func (s *Server) RegisterRoutes() {
	s.router.Route("/v1", func(r chi.Router) {
		r.Get("/commits", s.GetCommits)
		r.Get("/leaderboard", s.GetLeaderBoard)
	})

	s.router.NotFound(s.NotFoundHandler)
}
