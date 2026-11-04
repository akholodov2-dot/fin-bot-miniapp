package service

import (
	"errors"
	"fmt"
	"sync"

	"fin-bot-miniapp/internal/logger"
	"fin-bot-miniapp/internal/model"
	"fin-bot-miniapp/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCategoryCodeExists = errors.New("категория с таким кодом уже существует")
)

type CategoryService struct {
	repo    repository.CategoryRepository
	cache   map[int64][]*model.Category
	cacheMu sync.RWMutex
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		repo:  repo,
		cache: make(map[int64][]*model.Category),
	}
}

func (s *CategoryService) loadCategoriesFromDB(userID int64) error {
	categories, err := s.repo.GetActiveByUserID(userID)
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		return ErrCategoryNotFound
	}

	s.cacheMu.Lock()
	s.cache[userID] = categories
	s.cacheMu.Unlock()

	logger.Info("Loaded %d categories for user %d from database", len(categories), userID)
	return nil
}

func (s *CategoryService) initializeDefaultCategories(userID int64) {
	defaultCategories := []struct {
		code      string
		name      string
		emoji     string
		sortOrder int
	}{
		{"food", "Еда", "🍔", 1},
		{"shopping", "Покупки", "🛒", 2},
		{"transport", "Транспорт", "🚇", 3},
		{"housing", "Жильё", "🏠", 4},
		{"entertainment", "Развлечения", "🎬", 5},
		{"medical", "Здоровье", "🏥", 6},
		{"education", "Образование", "📚", 7},
		{"other", "Другое", "💰", 8},
	}

	for _, cat := range defaultCategories {
		category := &model.Category{
			UserID:    userID,
			Code:      cat.code,
			Name:      cat.name,
			Emoji:     cat.emoji,
			SortOrder: cat.sortOrder,
			Active:    true,
		}

		if err := s.repo.Upsert(category); err != nil {
			logger.Error("Failed to initialize category %s for user %d: %v", cat.code, userID, err)
		}
	}

	logger.Info("Initialized default categories for user %d", userID)

	if err := s.loadCategoriesFromDB(userID); err != nil {
		logger.Error("Failed to reload categories after initialization: %v", err)
	}
}

func (s *CategoryService) GetActiveByUserID(userID int64) ([]*model.Category, error) {
	s.cacheMu.RLock()
	cached, exists := s.cache[userID]
	s.cacheMu.RUnlock()

	if exists {
		result := make([]*model.Category, len(cached))
		copy(result, cached)
		return result, nil
	}

	if err := s.loadCategoriesFromDB(userID); err != nil {
		if err == ErrCategoryNotFound {
			s.initializeDefaultCategories(userID)

			s.cacheMu.RLock()
			cached, exists = s.cache[userID]
			s.cacheMu.RUnlock()

			if exists {
				result := make([]*model.Category, len(cached))
				copy(result, cached)
				return result, nil
			}
		}
		return nil, err
	}

	s.cacheMu.RLock()
	cached, _ = s.cache[userID]
	s.cacheMu.RUnlock()

	result := make([]*model.Category, len(cached))
	copy(result, cached)
	return result, nil
}

func (s *CategoryService) GetByUserIDAndCode(userID int64, code string) (*model.Category, error) {
	categories, err := s.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	for _, cat := range categories {
		if cat.Code == code {
			return cat, nil
		}
	}

	return nil, ErrCategoryNotFound
}

func (s *CategoryService) invalidateCache(userID int64) {
	s.cacheMu.Lock()
	delete(s.cache, userID)
	s.cacheMu.Unlock()
}

func (s *CategoryService) AddCategory(userID int64, code, name, emoji string, sortOrder int) error {
	_, err := s.repo.GetByUserIDAndCode(userID, code)
	if err == nil {
		return ErrCategoryCodeExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check existing category: %w", err)
	}

	category := &model.Category{
		UserID:    userID,
		Code:      code,
		Name:      name,
		Emoji:     emoji,
		SortOrder: sortOrder,
		Active:    true,
	}

	if err := s.repo.Upsert(category); err != nil {
		return err
	}

	logger.Info("Added new category: %s (%s) for user %d", name, code, userID)

	s.invalidateCache(userID)
	return nil
}
