package db

import (
	"database/sql"
	"errors"
)

// Task — модель задачи в базе данных.
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Создание / список

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

// Работа с одной записью

func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	task := &Task{}
	err := Db.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("The task was not found.")
		default:
			return nil, err
		}
	}
	return task, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := Db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	// метод RowsAffected() возвращает количество записей к которым
	// была применена SQL команда
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("Задача не найдена")
	}
	return nil
}

// DeleteTask — удаляет задачу по ID
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := Db.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("Задача не найдена")
	}
	return nil
}

// UpdateDate — обновляет дату при выполнение задачи
func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := Db.Exec(query, next, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("Задача не найдена")
	}
	return nil
}

// Поиск

// SearchByDate возвращает список задач, назначенных на конкретную дату.
// data — дата в формате ГГГГММДД (20060102);
// limit — максимальное количество возвращаемых записей.
func SearchByDate(data string, limit int) ([]*Task, error) {
	// Создаём пустой (но не nil) слайс.
	tasks := make([]*Task, 0)

	// Выбираем все колонки задач, у которых дата совпадает с указанной.
	// LIMIT ? ограничивает выборку.
	query := `SELECT * FROM scheduler WHERE date = ? LIMIT ?`

	// Выполняем запрос. На место "?" подставятся значения data и limit.
	rows, err := Db.Query(query, data, limit)
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

// SearchByText возвращает список задач, в заголовке или комментарии которых
// встречается подстрока data.
// data — текст для поиска;
// limit — максимальное количество возвращаемых записей.
func SearchByText(data string, limit int) ([]*Task, error) {
	// Создаём пустой (но не nil) слайс.
	tasks := make([]*Task, 0)

	// Выбираем задачи, где data встречается в заголовке ИЛИ комментарии.
	// LIKE ищет подстроку, а не точное совпадение.
	query := `SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? LIMIT ?`

	// Оборачиваем текст в символы % — так LIKE ищет вхождение в любом месте строки.
	like := "%" + data + "%"

	// Передаём одно и то же значение в оба "?" (для title и comment),
	// а последним аргументом — лимит.
	rows, err := Db.Query(query, like, like, limit)
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
