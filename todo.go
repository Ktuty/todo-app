package todo

import (
	"errors"
	"time"
)

type TodoList struct {
	Id          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Description string    `json:"description" db:"description"`
	Archived    bool      `json:"archived" db:"archived"`     // Новое поле для v2
	CreatedAt   time.Time `json:"created_at" db:"created_at"` // Новое поле для v2
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"` // Новое поле для v2
	Color       string    `json:"color,omitempty" db:"color"` // Новое поле в v2
	Priority    int       `json:"priority" db:"priority"`     // Новое поле в v2
}

type UsersList struct {
	Id     int `db:"id"`
	UserId int `db:"user_id"`
	ListId int `db:"list_id"`
}

type TodoItem struct {
	Id          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Description string    `json:"description" db:"description"`
	Done        bool      `json:"done" db:"done"`
	Archived    bool      `json:"archived" db:"archived"`     // Новое поле для v2
	CreatedAt   time.Time `json:"created_at" db:"created_at"` // Новое поле для v2
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"` // Новое поле для v2
}

type ListsItem struct {
	Id     int `db:"id"`
	ListId int `db:"list_id"`
	ItemId int `db:"item_id"`
}

// ListV2 - алиас для обратной совместимости, используйте TodoList вместо этого
type ListV2 struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Color       string    `json:"color,omitempty"`
	Priority    int       `json:"priority"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

type UpdateListInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Archived    *bool   `json:"archived"` // Добавлено для v2
	Color       *string `json:"color"`    // Добавлено для v2
	Priority    *int    `json:"priority"` // Добавлено для v2
}

func (i *UpdateListInput) Validate() error {
	if i.Title == nil && i.Description == nil && i.Archived == nil && i.Color == nil && i.Priority == nil {
		return errors.New("update structure has no values")
	}
	return nil
}

type UpdateItemInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done"`
	Archived    *bool   `json:"archived"` // Добавлено для v2
}

func (i *UpdateItemInput) Validate() error {
	if i.Title == nil && i.Description == nil && i.Done == nil && i.Archived == nil {
		return errors.New("update structure has no values")
	}
	return nil
}

type User struct {
	Id       int    `json:"-" db:"id"`
	Name     string `json:"name" binding:"required" db:"name"`
	Username string `json:"username" binding:"required" db:"username"`
	Password string `json:"password" binding:"required" db:"password_hash"`
}
