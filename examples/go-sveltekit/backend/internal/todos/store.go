package todos

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("todo not found")

type Todo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ownerSubject string
}

type Store struct {
	mutex sync.RWMutex
	todos map[string]Todo
}

func NewStore() *Store {
	return &Store{
		todos: make(map[string]Todo),
	}
}

func (s *Store) List(ownerSubject string) []Todo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]Todo, 0)

	for _, todo := range s.todos {
		if todo.ownerSubject == ownerSubject {
			result = append(result, todo)
		}
	}

	sort.Slice(result, func(first, second int) bool {
		return result[first].CreatedAt.Before(
			result[second].CreatedAt,
		)
	})

	return result
}

func (s *Store) Create(
	ownerSubject string,
	title string,
) Todo {
	now := time.Now().UTC()

	todo := Todo{
		ID:           uuid.NewString(),
		Title:        title,
		Completed:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
		ownerSubject: ownerSubject,
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.todos[todo.ID] = todo
	return todo
}

func (s *Store) Update(
	ownerSubject string,
	id string,
	title string,
	completed bool,
) (Todo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	todo, exists := s.todos[id]
	if !exists || todo.ownerSubject != ownerSubject {
		return Todo{}, ErrNotFound
	}

	todo.Title = title
	todo.Completed = completed
	todo.UpdatedAt = time.Now().UTC()

	s.todos[id] = todo
	return todo, nil
}

func (s *Store) Delete(
	ownerSubject string,
	id string,
) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	todo, exists := s.todos[id]
	if !exists || todo.ownerSubject != ownerSubject {
		return ErrNotFound
	}

	delete(s.todos, id)
	return nil
}
