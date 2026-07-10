package repository

func (r *InMemoryRepositoryURL) Reset(){
	clear(r.stor)
	r.filePath = ""
}
