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
	ErrLedgerNotFound  = errors.New("ledger not found")
	ErrLedgerCodeExists = errors.New("счёт с таким кодом уже существует")
)

type LedgerService struct {
	repo    repository.LedgerRepository
	cache   map[int64][]*model.Ledger
	cacheMu sync.RWMutex
}

func NewLedgerService(repo repository.LedgerRepository) *LedgerService {
	return &LedgerService{
		repo:  repo,
		cache: make(map[int64][]*model.Ledger),
	}
}

func (s *LedgerService) GetActiveByUserID(userID int64) ([]*model.Ledger, error) {
	s.cacheMu.RLock()
	cached, exists := s.cache[userID]
	s.cacheMu.RUnlock()

	if exists {
		result := make([]*model.Ledger, len(cached))
		copy(result, cached)
		return result, nil
	}

	ledgers, err := s.repo.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	if len(ledgers) == 0 {
		if err := s.initializeDefaultLedger(userID); err != nil {
			logger.Error("Failed to initialize default ledger for user %d: %v", userID, err)
			return nil, err
		}
		ledgers, err = s.repo.GetActiveByUserID(userID)
		if err != nil {
			return nil, err
		}
	}

	s.cacheMu.Lock()
	s.cache[userID] = ledgers
	s.cacheMu.Unlock()

	result := make([]*model.Ledger, len(ledgers))
	copy(result, ledgers)
	return result, nil
}

func (s *LedgerService) GetDefaultByUserID(userID int64) (*model.Ledger, error) {
	ledgers, err := s.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	for _, ledger := range ledgers {
		if ledger.IsDefault {
			return ledger, nil
		}
	}

	if len(ledgers) > 0 {
		return ledgers[0], nil
	}

	return nil, ErrLedgerNotFound
}

func (s *LedgerService) GetByID(id uint) (*model.Ledger, error) {
	return s.repo.GetByID(id)
}

func (s *LedgerService) AddLedger(userID int64, code, name, emoji string, sortOrder int) error {
	_, err := s.repo.GetByUserIDAndCode(userID, code)
	if err == nil {
		return ErrLedgerCodeExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check existing ledger: %w", err)
	}

	ledger := &model.Ledger{
		UserID:    userID,
		Code:      code,
		Name:      name,
		Emoji:     emoji,
		SortOrder: sortOrder,
		Active:    true,
		IsDefault: false,
	}

	if err := s.repo.Create(ledger); err != nil {
		return fmt.Errorf("create ledger: %w", err)
	}

	logger.Info("Added new ledger: %s (%s) for user %d", name, code, userID)

	s.invalidateCache(userID)
	return nil
}

func (s *LedgerService) SetDefault(userID int64, ledgerID uint) error {
	if err := s.repo.SetDefault(userID, ledgerID); err != nil {
		return fmt.Errorf("set default ledger: %w", err)
	}

	logger.Info("Set default ledger %d for user %d", ledgerID, userID)
	s.invalidateCache(userID)
	return nil
}

func (s *LedgerService) initializeDefaultLedger(userID int64) error {
	defaultLedger := &model.Ledger{
		UserID:    userID,
		Code:      "default",
		Name:      "Основной",
		Emoji:     "📒",
		SortOrder: 1,
		Active:    true,
		IsDefault: true,
	}

	if err := s.repo.Create(defaultLedger); err != nil {
		logger.Error("Failed to initialize default ledger for user %d: %v", userID, err)
		return err
	}

	logger.Info("Initialized default ledger for user %d", userID)
	return nil
}

func (s *LedgerService) invalidateCache(userID int64) {
	s.cacheMu.Lock()
	delete(s.cache, userID)
	s.cacheMu.Unlock()
}
