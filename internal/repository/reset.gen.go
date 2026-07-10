package repository

func (s *InMemoryRepositoryURL) Reset() {
	clear(s.stor)
	s.filePath = ""
}
