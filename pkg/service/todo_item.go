package service

import (
	"github.com/ktuty/todo-app"
	"github.com/ktuty/todo-app/pkg/repository"
)

type TodoItemService struct {
	repo     repository.TodoItem
	listRepo repository.TodoList
}

func NewTodoItemService(repo repository.TodoItem, listRepo repository.TodoList) *TodoItemService {
	return &TodoItemService{repo: repo, listRepo: listRepo}
}

func (s *TodoItemService) Create(userId, listId int, item todo.TodoItem) (int, error) {
	_, err := s.listRepo.GetById(userId, listId)
	if err != nil {
		// list does not exists or does not belongs to user
		return 0, err
	}

	return s.repo.Create(listId, item)
}

func (s *TodoItemService) GetAll(userId, listId int) ([]todo.TodoItem, error) {
	return s.repo.GetAll(userId, listId)
}

func (s *TodoItemService) GetById(userId, itemId int) (todo.TodoItem, error) {
	return s.repo.GetById(userId, itemId)
}

func (s *TodoItemService) Delete(userId, itemId int) error {
	return s.repo.Delete(userId, itemId)
}

func (s *TodoItemService) Update(userId, itemId int, input todo.UpdateItemInput) error {
	return s.repo.Update(userId, itemId, input)
}

// V2 методы

func (s *TodoItemService) GetAllWithPagination(userId, listId, offset, limit int, completed string) ([]todo.TodoItem, int, error) {
	// Временная реализация
	items, err := s.repo.GetAll(userId, listId)
	if err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию
	start := offset
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	if start >= len(items) {
		return []todo.TodoItem{}, len(items), nil
	}

	return items[start:end], len(items), nil
}

func (s *TodoItemService) ArchiveItem(userId, itemId int) error {
	// Временная реализация
	return s.repo.Delete(userId, itemId)
}

func (s *TodoItemService) CompleteItem(userId, itemId int) error {
	// Временная реализация
	done := true
	updateInput := todo.UpdateItemInput{
		Done: &done,
	}
	return s.repo.Update(userId, itemId, updateInput)
}
