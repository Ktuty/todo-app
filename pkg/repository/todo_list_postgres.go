package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ktuty/todo-app"
	"github.com/sirupsen/logrus"
)

type TodoListPostgres struct {
	db *sqlx.DB
}

func NewTodoListPostgres(db *sqlx.DB) *TodoListPostgres {
	return &TodoListPostgres{db: db}
}

func (r *TodoListPostgres) Create(userId int, list todo.TodoList) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createListQuery := fmt.Sprintf(`
		INSERT INTO %s (title, description, archived, created_at, updated_at, color, priority) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id`, todoListsTable)

	now := time.Now()
	row := tx.QueryRow(createListQuery,
		list.Title,
		list.Description,
		false, // archived по умолчанию false
		now,   // created_at
		now,   // updated_at
		list.Color,
		list.Priority)

	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	createUsersListQuery := fmt.Sprintf("INSERT INTO %s (user_id, list_id) VALUES ($1, $2)", usersListsTable)
	_, err = tx.Exec(createUsersListQuery, userId, id)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *TodoListPostgres) GetAll(userId int) ([]todo.TodoList, error) {
	var lists []todo.TodoList

	query := fmt.Sprintf(`
		SELECT tl.id, tl.title, tl.description, tl.archived, tl.created_at, tl.updated_at, tl.color, tl.priority 
		FROM %s tl 
		INNER JOIN %s ul on tl.id = ul.list_id 
		WHERE ul.user_id = $1 AND tl.archived = false`,
		todoListsTable, usersListsTable)
	err := r.db.Select(&lists, query, userId)

	return lists, err
}

func (r *TodoListPostgres) GetById(userId, listId int) (todo.TodoList, error) {
	var list todo.TodoList

	query := fmt.Sprintf(`
		SELECT tl.id, tl.title, tl.description, tl.archived, tl.created_at, tl.updated_at, tl.color, tl.priority 
		FROM %s tl
		INNER JOIN %s ul on tl.id = ul.list_id 
		WHERE ul.user_id = $1 AND ul.list_id = $2`,
		todoListsTable, usersListsTable)
	err := r.db.Get(&list, query, userId, listId)

	return list, err
}

func (r *TodoListPostgres) Delete(userId, listId int) error {
	query := fmt.Sprintf(`
		DELETE FROM %s tl 
		USING %s ul 
		WHERE tl.id = ul.list_id AND ul.user_id=$1 AND ul.list_id=$2`,
		todoListsTable, usersListsTable)
	_, err := r.db.Exec(query, userId, listId)

	return err
}

func (r *TodoListPostgres) Update(userId, listId int, input todo.UpdateListInput) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Title != nil {
		setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
		args = append(args, *input.Title)
		argId++
	}

	if input.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description=$%d", argId))
		args = append(args, *input.Description)
		argId++
	}

	if input.Archived != nil {
		setValues = append(setValues, fmt.Sprintf("archived=$%d", argId))
		args = append(args, *input.Archived)
		argId++
	}

	if input.Color != nil {
		setValues = append(setValues, fmt.Sprintf("color=$%d", argId))
		args = append(args, *input.Color)
		argId++
	}

	if input.Priority != nil {
		setValues = append(setValues, fmt.Sprintf("priority=$%d", argId))
		args = append(args, *input.Priority)
		argId++
	}

	// Всегда обновляем updated_at
	setValues = append(setValues, fmt.Sprintf("updated_at=$%d", argId))
	args = append(args, time.Now())
	argId++

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(`
		UPDATE %s tl SET %s 
		FROM %s ul 
		WHERE tl.id = ul.list_id AND ul.list_id=$%d AND ul.user_id=$%d`,
		todoListsTable, setQuery, usersListsTable, argId, argId+1)
	args = append(args, listId, userId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)
	return err
}

// ArchiveList - мягкое удаление списка
func (r *TodoListPostgres) ArchiveList(userId, listId int) error {
	query := fmt.Sprintf(`
		UPDATE %s tl SET archived = true, updated_at = $1 
		FROM %s ul 
		WHERE tl.id = ul.list_id AND ul.user_id=$2 AND ul.list_id=$3`,
		todoListsTable, usersListsTable)

	_, err := r.db.Exec(query, time.Now(), userId, listId)
	return err
}

// GetAllWithPagination получает списки с пагинацией
func (r *TodoListPostgres) GetAllWithPagination(userId, offset, limit int, archived string) ([]todo.TodoList, int, error) {
	var lists []todo.TodoList

	// Базовый запрос
	baseQuery := fmt.Sprintf(`
		SELECT tl.id, tl.title, tl.description, tl.archived, tl.created_at, tl.updated_at, tl.color, tl.priority 
		FROM %s tl 
		INNER JOIN %s ul on tl.id = ul.list_id 
		WHERE ul.user_id = $1`,
		todoListsTable, usersListsTable)

	// Добавляем фильтр по archived если указан
	query := baseQuery
	args := []interface{}{userId}
	argId := 2

	if archived != "" {
		isArchived := archived == "true"
		query += fmt.Sprintf(" AND tl.archived = $%d", argId)
		args = append(args, isArchived)
		argId++
	} else {
		// По умолчанию показываем только неархивированные
		query += " AND tl.archived = false"
	}

	// Добавляем сортировку и пагинацию
	query += " ORDER BY tl.priority DESC, tl.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, limit, offset)

	err := r.db.Select(&lists, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Получаем общее количество для пагинации
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s tl 
		INNER JOIN %s ul on tl.id = ul.list_id 
		WHERE ul.user_id = $1`,
		todoListsTable, usersListsTable)

	if archived != "" {
		//isArchived := archived == "true"
		countQuery += " AND tl.archived = $2"
	}

	var total int
	if archived != "" {
		err = r.db.Get(&total, countQuery, userId, archived == "true")
	} else {
		countQuery += " AND tl.archived = false"
		err = r.db.Get(&total, countQuery, userId)
	}

	if err != nil {
		return nil, 0, err
	}

	return lists, total, nil
}

// GetItemCount получает количество items в списке
func (r *TodoListPostgres) GetItemCount(userId, listId int) (int, error) {
	var count int
	query := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s ti 
		INNER JOIN %s li on li.item_id = ti.id
		INNER JOIN %s ul on ul.list_id = li.list_id 
		WHERE ul.user_id = $1 AND li.list_id = $2 AND ti.archived = false`,
		todoItemsTable, listsItemsTable, usersListsTable)

	err := r.db.Get(&count, query, userId, listId)
	return count, err
}
