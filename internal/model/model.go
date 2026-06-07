package model

type ShortenData struct {
	ID        string
	OriginURL string
}

type ShortenModel struct {
	ID          string
	OriginURL   string
	UserID      string
	DeletedFlag bool
}

type DeleteTaskDto struct {
	UserID string
	IDs    []string
}
