package chat

// navigationState is the former 3권 app-local navigation extension.
type navigationState struct{ pages []string }

func (s *navigationState) Back() bool {
	if len(s.pages) <= 1 {
		return false
	}
	s.pages = s.pages[:len(s.pages)-1]
	return true
}
