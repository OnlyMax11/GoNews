package model

// Post представляет новостную публикацию
type Post struct {
	ID      int    `json:"ID"`
	Title   string `json:"title"`
	Content string `json:"content"`
	PubTime int64  `json:"pubTime"`
	Link    string `json:"link"`
}
