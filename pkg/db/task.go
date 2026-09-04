package db

// Task — модель задачи в базе данных.
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask сохраняет новую задачу в таблицу scheduler
// и возвращает её идентификатор.
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := Db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	// Получаем идентификатор, присвоенный базой данных.
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
