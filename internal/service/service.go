package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"links-checker/internal/domain"
	"links-checker/internal/pdf"
	"links-checker/internal/service/checker"
)

const maxConcurrentChecks = 10

type Service struct {
	repo    Repository
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
			// добавляем в unknown статусе, чтобы потенциальный будущий робот мог возобновлять те проверки, которые почему-то оборвались
			Status: domain.StatusUnknown,
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

func (s *Service) GeneratePDFReportForTaskIds(ids []int64) ([]byte, error) {
	tasks := make([]*domain.LinkCheckTask, 0, len(ids))

	// По результатам задачи кажется достаточно печатать результаты как есть, то есть даже те unknown,
	// которые ещё не успели обработаться.
	// То есть предполагаю, что на лету запускать проверку или ещё что-то делать не нужно пока что
	for _, id := range ids {
		task, err := s.repo.GetLinkCheck(id)
		if err != nil {
			log.Printf("Failed to get task %d: %v", id, err)
			continue
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return []byte{}, errors.New("No tasks found for report generation")
	}

	return pdf.GenerateReport(tasks)
}
