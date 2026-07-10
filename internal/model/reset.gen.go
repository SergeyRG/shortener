package model

func (s *ShortenModel) Reset(){
	s.ID = ""
	s.OriginURL = ""
	s.UserID = ""
	s.DeletedFlag = false
}
