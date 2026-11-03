package online

// limitForTest ends the running game as if its time limit had been reached.
// It goes through the run loop, so it never races the game's own timers.
func (s *Server) limitForTest() {
	s.post(serverEvent{timeLimit: true})
}
