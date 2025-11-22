package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"links-checker/internal/domain"

	_ "github.com/mattn/go-sqlite3"
)

type Repository interface {
	SaveLinkCheckTask(check *domain.LinkCheckTask) (int64, error)
	GetLinkCheckTask(id int64) (*domain.LinkCheckTask, error)
	UpdateLinkStatus(checkID int64, url string, status domain.LinkStatus) error
}

type sqliteRepository struct {
	db *sql.DB
}

func New(dbPath string) (Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database err: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database err: %w", err)
	}

	repo := &sqliteRepository{db: db}
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("init schema err: %w", err)
	}

	return repo, nil
}

func (r *sqliteRepository) initSchema() error {
	// тут 2 таблицы: в одну просто складываем inbox сообщения (задачи на проверку), во вторую складываем статусы проверок каждого урла
	schema := `
	CREATE TABLE IF NOT EXISTS link_check_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		urls TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS link_statuses (
		task_id INTEGER NOT NULL,
		url TEXT NOT NULL,
		status TEXT NOT NULL,
		PRIMARY KEY (task_id, url),
		FOREIGN KEY (task_id) REFERENCES link_check_tasks(id)
	);
	`
	_, err := r.db.Exec(schema)
	return err
}

func (r *sqliteRepository) SaveLinkCheckTask(check *domain.LinkCheckTask) (int64, error) {
	//по хорошему надо переделать на использование моделей слоя БД, но т.к. задание тестовое, то кажется это не нужно
	urls := make([]string, len(check.Links))
	for i, link := range check.Links {
		urls[i] = link.URL
	}
	sort.Strings(urls)

	urlsJSON, err := json.Marshal(urls)
	if err != nil {
		return 0, fmt.Errorf("SaveLinkCheck marshal urls fail: %w", err)
	}

	// сначала проверим, может идентичное задание уже добавлялось и достаточно вернуть его ID
	var existingID int64
	err = r.db.QueryRow("SELECT id FROM link_check_tasks WHERE urls = ?", string(urlsJSON)).Scan(&existingID)
	if err == nil {
		return existingID, nil
	} else if err != sql.ErrNoRows {
		// если случалась какая-то другая ошибка, то мы не можем двигаться дальше, т.к. не знаем наверняка есть ли дубль задачи в БД
		return 0, fmt.Errorf("check existing links fail: %w", err)
	}

	// так как таблицы две и в одну нужно положить задачу, а во вторую её части, то обязательно в рамках транзакции сделать
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction fail: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"INSERT INTO link_check_tasks (urls, created_at) VALUES (?, ?)",
		string(urlsJSON), check.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert link check task fail: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id fail: %w", err)
	}

	for _, link := range check.Links {
		_, err = tx.Exec(
			"INSERT INTO link_statuses (task_id, url, status) VALUES (?, ?, ?)",
			id, link.URL, link.Status,
		)
		if err != nil {
			return 0, fmt.Errorf("insert link status fail: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction fail: %w", err)
	}

	return id, nil
}

func (r *sqliteRepository) GetLinkCheckTask(id int64) (*domain.LinkCheckTask, error) {
	var check domain.LinkCheckTask

	err := r.db.QueryRow(
		"SELECT id, created_at FROM link_check_tasks WHERE id = ?",
		id,
	).Scan(&check.ID, &check.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("link check task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get link check task fail: %w", err)
	}

	// для указанной задачи получаем результаты обработок всех ссылок
	rows, err := r.db.Query(
		"SELECT url, status FROM link_statuses WHERE task_id = ?",
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get links statuses error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var link domain.Link
		if err := rows.Scan(&link.URL, &link.Status); err != nil {
			return nil, fmt.Errorf("scan link status error: %w", err)
		}
		check.Links = append(check.Links, link)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating link statuses: %w", err)
	}

	return &check, nil
}

func (r *sqliteRepository) UpdateLinkStatus(checkID int64, url string, status domain.LinkStatus) error {
	result, err := r.db.Exec(
		"UPDATE link_statuses SET status = ? WHERE task_id = ? AND url = ?",
		status, checkID, url,
	)
	if err != nil {
		return fmt.Errorf("update link status fail: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected fail: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("link not found in task")
	}

	return nil
}
