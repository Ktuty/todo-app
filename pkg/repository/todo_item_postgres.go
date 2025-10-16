package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ktuty/todo-app"
)

type TodoItemPostgres struct {
	db *sqlx.DB
}

func NewTodoItemPostgres(db *sqlx.DB) *TodoItemPostgres {
	return &TodoItemPostgres{db: db}
}

func (r *TodoItemPostgres) Create(listId int, item todo.TodoItem) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var itemId int
	createItemQuery := fmt.Sprintf(`
		INSERT INTO %s (title, description, done, archived, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id`, todoItemsTable)

	now := time.Now()
	row := tx.QueryRow(createItemQuery,
		item.Title,
		item.Description,
		false, // done по умолчанию false
		false, // archived по умолчанию false
		now,   // created_at
		now)   // updated_at

	err = row.Scan(&itemId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	createListItemsQuery := fmt.Sprintf("INSERT INTO %s (list_id, item_id) VALUES ($1, $2)", listsItemsTable)
	_, err = tx.Exec(createListItemsQuery, listId, itemId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return itemId, tx.Commit()
}

func (r *TodoItemPostgres) GetAll(userId, listId int) ([]todo.TodoItem, error) {
	var items []todo.TodoItem
	query := fmt.Sprintf(`
		SELECT ti.id, ti.title, ti.description, ti.done, ti.archived, ti.created_at, ti.updated_at 
		FROM %s ti 
		INNER JOIN %s li on li.item_id = ti.id
		INNER JOIN %s ul on ul.list_id = li.list_id 
		WHERE li.list_id = $1 AND ul.user_id = $2 AND ti.archived = false`,
		todoItemsTable, listsItemsTable, usersListsTable)
	if err := r.db.Select(&items, query, listId, userId); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *TodoItemPostgres) GetById(userId, itemId int) (todo.TodoItem, error) {
	var item todo.TodoItem
	query := fmt.Sprintf(`
		SELECT ti.id, ti.title, ti.description, ti.done, ti.archived, ti.created_at, ti.updated_at 
		FROM %s ti 
		INNER JOIN %s li on li.item_id = ti.id
		INNER JOIN %s ul on ul.list_id = li.list_id 
		WHERE ti.id = $1 AND ul.user_id = $2`,
		todoItemsTable, listsItemsTable, usersListsTable)
	if err := r.db.Get(&item, query, itemId, userId); err != nil {
		return item, err
	}

	return item, nil
}

func (r *TodoItemPostgres) Delete(userId, itemId int) error {
	query := fmt.Sprintf(`
		DELETE FROM %s ti 
		USING %s li, %s ul 
		WHERE ti.id = li.item_id AND li.list_id = ul.list_id AND ul.user_id = $1 AND ti.id = $2`,
		todoItemsTable, listsItemsTable, usersListsTable)
	_, err := r.db.Exec(query, userId, itemId)
	return err
}

func (r *TodoItemPostgres) Update(userId, itemId int, input todo.UpdateItemInput) error {
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

	if input.Done != nil {
		setValues = append(setValues, fmt.Sprintf("done=$%d", argId))
		args = append(args, *input.Done)
		argId++
	}

	if input.Archived != nil {
		setValues = append(setValues, fmt.Sprintf("archived=$%d", argId))
		args = append(args, *input.Archived)
		argId++
	}

	// Всегда обновляем updated_at
	setValues = append(setValues, fmt.Sprintf("updated_at=$%d", argId))
	args = append(args, time.Now())
	argId++

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(`
		UPDATE %s ti SET %s 
		FROM %s li, %s ul
		WHERE ti.id = li.item_id AND li.list_id = ul.list_id AND ul.user_id = $%d AND ti.id = $%d`,
		todoItemsTable, setQuery, listsItemsTable, usersListsTable, argId, argId+1)
	args = append(args, userId, itemId)

	_, err := r.db.Exec(query, args...)
	return err
}

// ArchiveItem - мягкое удаление item
func (r *TodoItemPostgres) ArchiveItem(userId, itemId int) error {
	query := fmt.Sprintf(`
		UPDATE %s ti SET archived = true, updated_at = $1 
		FROM %s li, %s ul
		WHERE ti.id = li.item_id AND li.list_id = ul.list_id AND ul.user_id = $2 AND ti.id = $3`,
		todoItemsTable, listsItemsTable, usersListsTable)

	_, err := r.db.Exec(query, time.Now(), userId, itemId)
	return err
}

// GetAllWithPagination получает items с пагинацией
func (r *TodoItemPostgres) GetAllWithPagination(userId, listId, offset, limit int, completed string) ([]todo.TodoItem, int, error) {
	var items []todo.TodoItem

	baseQuery := fmt.Sprintf(`
		SELECT ti.id, ti.title, ti.description, ti.done, ti.archived, ti.created_at, ti.updated_at 
		FROM %s ti 
		INNER JOIN %s li on li.item_id = ti.id
		INNER JOIN %s ul on ul.list_id = li.list_id 
		WHERE ul.user_id = $1 AND ti.archived = false`,
		todoItemsTable, listsItemsTable, usersListsTable)

	query := baseQuery
	args := []interface{}{userId}
	argId := 2

	if listId > 0 {
		query += fmt.Sprintf(" AND li.list_id = $%d", argId)
		args = append(args, listId)
		argId++
	}

	if completed != "" {
		isCompleted := completed == "true"
		query += fmt.Sprintf(" AND ti.done = $%d", argId)
		args = append(args, isCompleted)
		argId++
	}

	query += " ORDER BY ti.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, limit, offset)

	err := r.db.Select(&items, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Получаем общее количество
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s ti 
		INNER JOIN %s li on li.item_id = ti.id
		INNER JOIN %s ul on ul.list_id = li.list_id 
		WHERE ul.user_id = $1 AND ti.archived = false`,
		todoItemsTable, listsItemsTable, usersListsTable)

	countArgs := []interface{}{userId}
	countArgId := 2

	if listId > 0 {
		countQuery += fmt.Sprintf(" AND li.list_id = $%d", countArgId)
		countArgs = append(countArgs, listId)
		countArgId++
	}

	if completed != "" {
		isCompleted := completed == "true"
		countQuery += fmt.Sprintf(" AND ti.done = $%d", countArgId)
		countArgs = append(countArgs, isCompleted)
		countArgId++
	}

	var total int
	err = r.db.Get(&total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CompleteItem - отмечает item как выполненный
func (r *TodoItemPostgres) CompleteItem(userId, itemId int) error {
	query := fmt.Sprintf(`
		UPDATE %s ti SET done = true, updated_at = $1 
		FROM %s li, %s ul
		WHERE ti.id = li.item_id AND li.list_id = ul.list_id AND ul.user_id = $2 AND ti.id = $3`,
		todoItemsTable, listsItemsTable, usersListsTable)

	_, err := r.db.Exec(query, time.Now(), userId, itemId)
	return err
}
