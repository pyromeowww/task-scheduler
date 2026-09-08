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

// Tasks возвращает список задач из таблицы scheduler.
// Задачи сортируются по дате по возрастанию (сначала ближайшие),
// а количество возвращаемых записей ограничено параметром limit.
func Tasks(limit int) ([]*Task, error) {
	// Создаём пустой (но не nil) слайс.
	tasks := make([]*Task, 0)

	// Явно перечисляем колонки: порядок в SELECT должен совпадать
	// с порядком аргументов в rows.Scan ниже.
	// ORDER BY date сортирует задачи по дате, LIMIT ? ограничивает выборку.
	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`

	// Выполняем запрос. На место "?" подставится значение limit.
	rows, err := Db.Query(query, limit)
	// Если база не смогла выполнить запрос — сразу возвращаем ошибку.
	if err != nil {
		return tasks, err
	}
	// Гарантируем, что соединение с БД освободится после выхода из функции,
	// даже если произойдёт ошибка в середине цикла.
	defer rows.Close()

	// Проходим по всем строкам результата.
	// Next() переводит курсор на следующую строку и возвращает true,
	// пока строки ещё есть.
	for rows.Next() {
		// Создаём новую задачу для текущей строки.
		task := &Task{}
		// Копируем значения колонок текущей строки в поля task.
		// Адреса (&task.ID и т.д.) нужны, чтобы Scan мог записать данные внутрь.
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return tasks, err
		}
		// Добавляем прочитанную задачу в итоговый список.
		tasks = append(tasks, task)
	}
	// Проверяем, не прервался ли обход строк из-за ошибки.
	// Next() возвращает false либо когда строки закончились, либо при сбое
	// rows.Err() позволяет отличить одно от другого.
	if err := rows.Err(); err != nil {
		return tasks, err
	}
	return tasks, nil
}
