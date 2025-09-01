package usecases

import (
	Domain "lawgen/admin-service/Domain"
)

type FeedbackUsecase struct {
    repo Domain.FeedbackRepository
}

func NewFeedbackUsecase(r Domain.FeedbackRepository) *FeedbackUsecase {
    return &FeedbackUsecase{repo: r}
}

func (uc *FeedbackUsecase) CreateFeedback(fb *Domain.Feedback) error {
    return uc.repo.Create(fb)
}

func (uc *FeedbackUsecase) GetFeedbackByID(id string) (*Domain.Feedback, error) {
    return uc.repo.GetByID(id)
}

func (uc *FeedbackUsecase) ListFeedbacks() ([]Domain.Feedback, error) {
    return uc.repo.List()
}