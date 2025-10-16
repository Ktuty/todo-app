package service

import (
	"github.com/ktuty/todo-app"
	"github.com/ktuty/todo-app/pkg/repository"
)

type TodoListService struct {
	repo repository.TodoList
}

func NewTodoListService(repo repository.TodoList) *TodoListService {
	return &TodoListService{repo: repo}
}

func (s *TodoListService) Create(userId int, list todo.TodoList) (int, error) {
	return s.repo.Create(userId, list)
}

func (s *TodoListService) GetAll(userId int) ([]todo.TodoList, error) {
	return s.repo.GetAll(userId)
}

func (s *TodoListService) GetById(userId, listId int) (todo.TodoList, error) {
	return s.repo.GetById(userId, listId)
}

func (s *TodoListService) Delete(userId, listId int) error {
	return s.repo.Delete(userId, listId)
}

func (s *TodoListService) Update(userId, listId int, input todo.UpdateListInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return s.repo.Update(userId, listId, input)
}

// V2 методы

func (s *TodoListService) GetAllWithPagination(userId, offset, limit int, archived string) ([]todo.TodoList, int, error) {
	// Временная реализация - возвращаем все списки
	// В реальной реализации нужно добавить пагинацию в репозиторий
	lists, err := s.repo.GetAll(userId)
	if err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию
	start := offset
	if start > len(lists) {
		start = len(lists)
	}
	end := start + limit
	if end > len(lists) {
		end = len(lists)
	}

	if start >= len(lists) {
		return []todo.TodoList{}, len(lists), nil
	}

	return lists[start:end], len(lists), nil
}

func (s *TodoListService) GetItemCount(userId, listId int) (int, error) {
	// Временная реализация
	// В реальной реализации нужно добавить подсчет items в репозиторий
	return 0, nil
}

func (s *TodoListService) ArchiveList(userId, listId int) error {
	// Временная реализация - используем обычное удаление
	// В реальной реализации нужно добавить поле archived в базу
	return s.repo.Delete(userId, listId)
}
