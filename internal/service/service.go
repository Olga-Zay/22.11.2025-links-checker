package service

import (
	"context"
	"log"
	"sync"
	"time"

	"links-checker/internal/checker"
	"links-checker/internal/domain"
	"links-checker/internal/repository"
)

const maxConcurrentChecks = 10

type Service struct {
	repo    repository.Repository
	checker *checker.LinkChecker
}

type Repository interface {
	SaveLinkCheck(check *domain.LinkCheckTask) (int64, error)
	GetLinkCheck(id int64) (*domain.LinkCheckTask, error)
	UpdateLinkStatus(checkID int64, url string, status domain.LinkStatus) error
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:    repo,
		checker: checker.New(),
	}
}

func (s *Service) CheckLinks(urls []string) (int64, error) {
	links := make([]domain.Link, len(urls))
	for i, url := range urls {
		if url == "" {
			log.Println("Skipped empty url from list")
			continue
		}
		links[i] = domain.Link{
			URL: url,
			// добавляем в pending статусе, чтобы потенциальный будущий робот мог возобновлять те проверки, которые почему-то оборвались
			Status: domain.StatusPending,
		}
	}

	checkTask := &domain.LinkCheckTask{
		Links:     links,
		CreatedAt: time.Now(),
	}

	taskID, err := s.repo.SaveLinkCheck(checkTask)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	semaphore := make(chan bool, maxConcurrentChecks)

	for _, url := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()

			semaphore <- true
			defer func() {
				<-semaphore
			}()

			status := s.checker.Check(ctx, url)
			if err := s.repo.UpdateLinkStatus(taskID, url, status); err != nil {
				log.Printf("Failed to update status for  %s: %v", url, err)
			}
		}()
	}

	wg.Wait()

	return taskID, nil
}

func (s *Service) GetLinkCheckTaskResults(id int64) (*domain.LinkCheckTask, error) {
	return s.repo.GetLinkCheck(id)
}
