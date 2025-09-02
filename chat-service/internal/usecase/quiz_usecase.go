package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type quizUseCase struct {
	quizRepo domain.IQuizRepository
}

func NewQuizUseCase(quizRepo domain.IQuizRepository) domain.IQuizUseCase {
	return &quizUseCase{quizRepo: quizRepo}
}

// --- Category Methods ---

func (u *quizUseCase) CreateCategory(ctx context.Context, name string) (*domain.QuizCategory, error) {
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	category := &domain.QuizCategory{Name: name}
	err := u.quizRepo.CreateCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (u *quizUseCase) GetCategory(ctx context.Context, id string) (*domain.QuizCategory, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}
	return u.quizRepo.GetCategoryByID(ctx, objID)
}

func (u *quizUseCase) ListCategories(ctx context.Context, page, limit int64) (*domain.PaginatedQuizCategories, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	categories, total, err := u.quizRepo.GetCategories(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedQuizCategories{
		Items:       categories,
		TotalItems:  total,
		TotalPages:  int64(math.Ceil(float64(total) / float64(limit))),
		CurrentPage: page,
		PageSize:    limit,
	}, nil
}

func (u *quizUseCase) UpdateCategory(ctx context.Context, id, name string) (*domain.QuizCategory, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}
	if name == "" {
		return nil, errors.New("category name cannot be empty")
	}
	category, err := u.quizRepo.GetCategoryByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	category.Name = name
	err = u.quizRepo.UpdateCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (u *quizUseCase) DeleteCategory(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid category ID")
	}
	return u.quizRepo.DeleteCategory(ctx, objID)
}

// --- Quiz Methods ---

func (u *quizUseCase) CreateQuiz(ctx context.Context, categoryID, name, description string) (*domain.Quiz, error) {
	catObjID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}
	if name == "" {
		return nil, errors.New("quiz name cannot be empty")
	}
	quiz := &domain.Quiz{
		CategoryID:  catObjID,
		Name:        name,
		Description: description,
		Questions:   []domain.Question{},
	}
	err = u.quizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
	}
	return quiz, nil
}

func (u *quizUseCase) GetQuiz(ctx context.Context, id string) (*domain.Quiz, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid quiz ID")
	}
	return u.quizRepo.GetQuizByID(ctx, objID)
}

func (u *quizUseCase) ListQuizzesByCategory(ctx context.Context, categoryID string, page, limit int64) (*domain.PaginatedQuizzes, error) {
	catObjID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, errors.New("invalid category ID")
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	quizzes, total, err := u.quizRepo.GetQuizzesByCategoryID(ctx, catObjID, page, limit)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedQuizzes{
		Items:       quizzes,
		TotalItems:  total,
		TotalPages:  int64(math.Ceil(float64(total) / float64(limit))),
		CurrentPage: page,
		PageSize:    limit,
	}, nil
}

func (u *quizUseCase) UpdateQuiz(ctx context.Context, id, name, description string) (*domain.Quiz, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid quiz ID")
	}
	if name == "" {
		return nil, errors.New("quiz name cannot be empty")
	}
	quiz, err := u.quizRepo.GetQuizByID(ctx, objID)
	if err != nil {
		return nil, err
	}
	quiz.Name = name
	quiz.Description = description
	err = u.quizRepo.UpdateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
	}
	return quiz, nil
}

func (u *quizUseCase) DeleteQuiz(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid quiz ID")
	}
	return u.quizRepo.DeleteQuiz(ctx, objID)
}

// --- Question Methods ---

func (u *quizUseCase) AddQuestion(ctx context.Context, quizID string, text string, options map[string]string, correctOption string) (*domain.Quiz, error) {
	quizObjID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		return nil, errors.New("invalid quiz ID")
	}
	if text == "" {
		return nil, errors.New("question text cannot be empty")
	}
	if len(options) == 0 {
		return nil, errors.New("question must have options")
	}
	if _, ok := options[correctOption]; !ok {
		return nil, errors.New("correct option must be one of the provided options")
	}

	question := &domain.Question{
		Text:          text,
		Options:       options,
		CorrectOption: correctOption,
	}

	err = u.quizRepo.AddQuestionToQuiz(ctx, quizObjID, question)
	if err != nil {
		return nil, err
	}
	return u.quizRepo.GetQuizByID(ctx, quizObjID)
}

func (u *quizUseCase) UpdateQuestion(ctx context.Context, quizID, questionID, text string, options map[string]string, correctOption string) (*domain.Question, error) {
	quizObjID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		return nil, errors.New("invalid quiz ID")
	}
	questionObjID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return nil, errors.New("invalid question ID")
	}
	if text == "" {
		return nil, errors.New("question text cannot be empty")
	}
	if len(options) == 0 {
		return nil, errors.New("question must have options")
	}
	if _, ok := options[correctOption]; !ok {
		return nil, errors.New("correct option must be one of the provided options")
	}

	question, err := u.quizRepo.GetQuestionByID(ctx, quizObjID, questionObjID)
	if err != nil {
		return nil, err
	}

	question.Text = text
	question.Options = options
	question.CorrectOption = correctOption

	err = u.quizRepo.UpdateQuestionInQuiz(ctx, quizObjID, question)
	if err != nil {
		return nil, err
	}
	return question, nil
}

func (u *quizUseCase) DeleteQuestion(ctx context.Context, quizID, questionID string) error {
	quizObjID, err := primitive.ObjectIDFromHex(quizID)
	if err != nil {
		return errors.New("invalid quiz ID")
	}
	questionObjID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		return errors.New("invalid question ID")
	}
	return u.quizRepo.DeleteQuestionFromQuiz(ctx, quizObjID, questionObjID)
}
