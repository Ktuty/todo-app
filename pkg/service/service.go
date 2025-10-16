package service

import (
	"time"

	"github.com/ktuty/todo-app"
	"github.com/ktuty/todo-app/pkg/repository"
)

type Authorization interface {
	CreateUser(user todo.User) (int, error)
	GenerateTocken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type TodoList interface {
	Create(userId int, list todo.TodoList) (int, error)
	GetAll(userId int) ([]todo.TodoList, error)
	GetById(userId, listId int) (todo.TodoList, error)
	Delete(userId, listId int) error
	Update(userId, listId int, input todo.UpdateListInput) error
	// V2 методы
	GetAllWithPagination(userId, offset, limit int, archived string) ([]todo.TodoList, int, error)
	GetItemCount(userId, listId int) (int, error)
	ArchiveList(userId, listId int) error
}

type TodoItem interface {
	Create(userId, listId int, item todo.TodoItem) (int, error)
	GetAll(userId, listId int) ([]todo.TodoItem, error)
	GetById(userId, itemId int) (todo.TodoItem, error)
	Delete(userId, itemId int) error
	Update(userId, itemId int, input todo.UpdateItemInput) error
	// V2 методы
	GetAllWithPagination(userId, listId, offset, limit int, completed string) ([]todo.TodoItem, int, error)
	ArchiveItem(userId, itemId int) error
	CompleteItem(userId, itemId int) error
}

type Idempotency interface {
	CheckIdempotency(userId int, key string) (int, error)
	StoreIdempotency(userId int, key string, resourceId int, ttl time.Duration) error
}

type Service struct {
	Authorization
	TodoList
	TodoItem
	Idempotency
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		TodoList:      NewTodoListService(repos.TodoList),
		TodoItem:      NewTodoItemService(repos.TodoItem, repos.TodoList),
		Idempotency:   NewIdempotencyService(), // Добавляем сервис идемпотентности
	}
}
