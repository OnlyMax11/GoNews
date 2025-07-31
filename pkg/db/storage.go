package db

import (
	"database/sql"
	"log"

	"GoNews/pkg/model"

	_ "github.com/lib/pq"
)

// Storage представляет интерфейс к БД
type Storage struct {
	db *sql.DB
}

// New создает новое подключение к БД
func New(dsn string) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

// Close закрывает соединение с БД
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// func createTable(db *sql.DB) error {
// 	_, err := db.Exec(`
//         CREATE TABLE IF NOT EXISTS posts (
//             id SERIAL PRIMARY KEY,
//             title TEXT NOT NULL,
//             content TEXT NOT NULL,
//             pub_time BIGINT NOT NULL,
//             link TEXT NOT NULL UNIQUE,
//             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//         );
//         CREATE INDEX IF NOT EXISTS idx_pub_time ON posts(pub_time DESC);
//     `)
// 	return err
// }

// SavePost сохраняет публикацию в БД
func (s *Storage) SavePost(p model.Post) error {
	_, err := s.db.Exec(`
        INSERT INTO posts (title, content, pub_time, link)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (link) DO NOTHING`,
		p.Title, p.Content, p.PubTime, p.Link)
	return err
}

// GetPosts возвращает последние n публикаций
func (s *Storage) GetPosts(n int) ([]model.Post, error) {
	rows, err := s.db.Query(`
        SELECT id, title, content, pub_time, link 
        FROM posts 
        ORDER BY pub_time DESC 
        LIMIT $1`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.PubTime, &p.Link); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		posts = append(posts, p)
	}
	return posts, nil
}
