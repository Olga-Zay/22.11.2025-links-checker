package repository

import "links-checker/internal/domain"

type Repository interface {
	SaveLinkCheck(check *domain.LinkCheckTask) (int64, error)
	GetLinkCheck(id int64) (*domain.LinkCheckTask, error)
	UpdateLinkStatus(checkID int64, url string, status domain.LinkStatus) error
}
