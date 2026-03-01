package repository

type InMemoryRepositoryURL struct {
	stor map[string]string
}

func NewInMemoryRepositoryURL() *InMemoryRepositoryURL {
	return &InMemoryRepositoryURL{
		stor: make(map[string]string),
	}
}

func (r *InMemoryRepositoryURL) Add(url string, id string) {
	r.stor[id] = url
}

func (r *InMemoryRepositoryURL) GetByID(id string) (string, bool) {
	if v, ok := r.stor[id]; ok {
		return v, true
	} else {
		return "", false
	}

}
