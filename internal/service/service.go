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

type Service struct {
	repo    repository.Repository
	checker *checker.LinkChecker
}

func NewService(repo repository.Repository) *Service {
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
			URL:    url,
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

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			status := s.checker.Check(ctx, u)
			if err := s.repo.UpdateLinkStatus(taskID, u, status); err != nil {
				log.Printf("Failed to update status for %s: %v", u, err)
			}
		}(url)
	}

	wg.Wait()

	return taskID, nil
}

func (s *Service) GetLinkCheckTaskResults(id int64) (*domain.LinkCheckTask, error) {
	return s.repo.GetLinkCheck(id)
}
