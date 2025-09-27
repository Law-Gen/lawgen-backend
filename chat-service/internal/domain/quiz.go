package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Question struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Text          string             `bson:"text" json:"text"`
	Options       map[string]string  `bson:"options" json:"options"`
	CorrectOption string             `bson:"correct_option" json:"correct_option"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type Quiz struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID     primitive.ObjectID `bson:"category_id" json:"category_id"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description" json:"description"`
	Questions      []Question         `bson:"questions" json:"questions"`
	TotalQuestions int                `bson:"total_questions" json:"total_questions"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type QuizCategory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	TotalQuizzes int                `bson:"total_quizzes" json:"total_quizzes"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type PaginatedQuizCategories struct {
	Items       []*QuizCategory `json:"items"`
	TotalItems  int64           `json:"total_items"`
	TotalPages  int64           `json:"total_pages"`
	CurrentPage int64           `json:"current_page"`
	PageSize    int64           `json:"page_size"`
}

type PaginatedQuizzes struct {
	Items       []*Quiz `json:"items"`
	TotalItems  int64   `json:"total_items"`
	TotalPages  int64   `json:"total_pages"`
	CurrentPage int64   `json:"current_page"`
	PageSize    int64   `json:"page_size"`
}

type IQuizRepository interface {
	// Category methods
	CreateCategory(ctx context.Context, category *QuizCategory) error
	GetCategoryByID(ctx context.Context, id primitive.ObjectID) (*QuizCategory, error)
	GetCategories(ctx context.Context, page, limit int64) ([]*QuizCategory, int64, error)
	UpdateCategory(ctx context.Context, category *QuizCategory) error
	// delete the quizzes and their questions recursively
	DeleteCategory(ctx context.Context, id primitive.ObjectID) error

	// Quiz methods
	CreateQuiz(ctx context.Context, quiz *Quiz) error
	GetQuizByID(ctx context.Context, id primitive.ObjectID) (*Quiz, error)
	GetQuizzesByCategoryID(ctx context.Context, categoryID primitive.ObjectID, page, limit int64) ([]*Quiz, int64, error)
	UpdateQuiz(ctx context.Context, quiz *Quiz) error
	// delete the questions recursively
	DeleteQuiz(ctx context.Context, id primitive.ObjectID) error

	// Question methods
	AddQuestionToQuiz(ctx context.Context, quizID primitive.ObjectID, question *Question) error
	GetQuestionByID(ctx context.Context, quizID, questionID primitive.ObjectID) (*Question, error)
	UpdateQuestionInQuiz(ctx context.Context, quizID primitive.ObjectID, question *Question) error
	DeleteQuestionFromQuiz(ctx context.Context, quizID, questionID primitive.ObjectID) error
}

type IQuizUseCase interface {
	// Category methods
	CreateCategory(ctx context.Context, name string) (*QuizCategory, error)
	GetCategory(ctx context.Context, id string) (*QuizCategory, error)
	ListCategories(ctx context.Context, page, limit int64) (*PaginatedQuizCategories, error)
	UpdateCategory(ctx context.Context, id, name string) (*QuizCategory, error)
	DeleteCategory(ctx context.Context, id string) error

	// Quiz methods
	CreateQuiz(ctx context.Context, categoryID, name, description string) (*Quiz, error)
	GetQuiz(ctx context.Context, id string) (*Quiz, error)
	ListQuizzesByCategory(ctx context.Context, categoryID string, page, limit int64) (*PaginatedQuizzes, error)
	UpdateQuiz(ctx context.Context, id, name, description string) (*Quiz, error)
	DeleteQuiz(ctx context.Context, id string) error

	// Question methods
	AddQuestion(ctx context.Context, quizID string, text string, options map[string]string, correctOption string) (*Quiz, error)
	UpdateQuestion(ctx context.Context, quizID, questionID, text string, options map[string]string, correctOption string) (*Question, error)
	DeleteQuestion(ctx context.Context, quizID, questionID string) error
}
