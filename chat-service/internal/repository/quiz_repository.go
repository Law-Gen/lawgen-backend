package repository

import (
	"context"
	"errors"
	"time"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type quizRepository struct {
	db *mongo.Database
}

func NewQuizRepository(db *mongo.Database) domain.IQuizRepository {
	return &quizRepository{db: db}
}

func (r *quizRepository) quizCategoriesCollection() *mongo.Collection {
	return r.db.Collection("quiz_categories")
}

func (r *quizRepository) quizzesCollection() *mongo.Collection {
	return r.db.Collection("quizzes")
}

// --- Category Methods ---

func (r *quizRepository) CreateCategory(ctx context.Context, category *domain.QuizCategory) error {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	_, err := r.quizCategoriesCollection().InsertOne(ctx, category)
	return err
}

func (r *quizRepository) GetCategoryByID(ctx context.Context, id primitive.ObjectID) (*domain.QuizCategory, error) {
	var category domain.QuizCategory
	err := r.quizCategoriesCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

func (r *quizRepository) GetCategories(ctx context.Context, page, limit int64) ([]*domain.QuizCategory, int64, error) {
	findOptions := options.Find()
	findOptions.SetSkip((page - 1) * limit)
	findOptions.SetLimit(limit)

	cursor, err := r.quizCategoriesCollection().Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var categories []*domain.QuizCategory
	if err = cursor.All(ctx, &categories); err != nil {
		return nil, 0, err
	}

	total, err := r.quizCategoriesCollection().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *quizRepository) UpdateCategory(ctx context.Context, category *domain.QuizCategory) error {
	category.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":       category.Name,
			"updated_at": category.UpdatedAt,
		},
	}
	_, err := r.quizCategoriesCollection().UpdateOne(ctx, bson.M{"_id": category.ID}, update)
	return err
}

func (r *quizRepository) DeleteCategory(ctx context.Context, id primitive.ObjectID) error {
	// Also delete quizzes associated with this category
	_, err := r.quizzesCollection().DeleteMany(ctx, bson.M{"category_id": id})
	if err != nil {
		return err
	}
	// Reset total_quizzes before deleting the category
	_, err = r.quizCategoriesCollection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"total_quizzes": 0}})
	if err != nil {
		return err
	}
	_, err = r.quizCategoriesCollection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// --- Quiz Methods ---

func (r *quizRepository) CreateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	quiz.ID = primitive.NewObjectID()
	quiz.CreatedAt = time.Now()
	quiz.UpdatedAt = time.Now()
	quiz.TotalQuestions = len(quiz.Questions)
	_, err := r.quizzesCollection().InsertOne(ctx, quiz)
	if err != nil {
		return err
	}
	// Increment total_quizzes in category
	_, err = r.quizCategoriesCollection().UpdateOne(ctx, bson.M{"_id": quiz.CategoryID}, bson.M{"$inc": bson.M{"total_quizzes": 1}})
	return err
}

func (r *quizRepository) GetQuizByID(ctx context.Context, id primitive.ObjectID) (*domain.Quiz, error) {
	var quiz domain.Quiz
	err := r.quizzesCollection().FindOne(ctx, bson.M{"_id": id}).Decode(&quiz)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("quiz not found")
		}
		return nil, err
	}
	return &quiz, nil
}

func (r *quizRepository) GetQuizzesByCategoryID(ctx context.Context, categoryID primitive.ObjectID, page, limit int64) ([]*domain.Quiz, int64, error) {
	findOptions := options.Find()
	findOptions.SetSkip((page - 1) * limit)
	findOptions.SetLimit(limit)

	filter := bson.M{"category_id": categoryID}
	cursor, err := r.quizzesCollection().Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var quizzes []*domain.Quiz
	if err = cursor.All(ctx, &quizzes); err != nil {
		return nil, 0, err
	}

	total, err := r.quizzesCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return quizzes, total, nil
}

func (r *quizRepository) UpdateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	quiz.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":        quiz.Name,
			"description": quiz.Description,
			"updated_at":  quiz.UpdatedAt,
		},
	}
	_, err := r.quizzesCollection().UpdateOne(ctx, bson.M{"_id": quiz.ID}, update)
	return err
}

func (r *quizRepository) DeleteQuiz(ctx context.Context, id primitive.ObjectID) error {
	// Find the quiz to get its category
	quiz, err := r.GetQuizByID(ctx, id)
	if err != nil {
		return err
	}
	_, err = r.quizzesCollection().DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	// Decrement total_quizzes in category
	_, err = r.quizCategoriesCollection().UpdateOne(ctx, bson.M{"_id": quiz.CategoryID}, bson.M{"$inc": bson.M{"total_quizzes": -1}})
	return err
}

// --- Question Methods ---

func (r *quizRepository) AddQuestionToQuiz(ctx context.Context, quizID primitive.ObjectID, question *domain.Question) error {
	question.ID = primitive.NewObjectID()
	question.CreatedAt = time.Now()
	question.UpdatedAt = time.Now()
	update := bson.M{"$push": bson.M{"questions": question}, "$inc": bson.M{"total_questions": 1}}
	_, err := r.quizzesCollection().UpdateOne(ctx, bson.M{"_id": quizID}, update)
	return err
}

func (r *quizRepository) GetQuestionByID(ctx context.Context, quizID, questionID primitive.ObjectID) (*domain.Question, error) {
	var quiz domain.Quiz
	filter := bson.M{"_id": quizID, "questions._id": questionID}
	projection := bson.M{"questions.$": 1}
	err := r.quizzesCollection().FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&quiz)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("question not found")
		}
		return nil, err
	}
	if len(quiz.Questions) == 0 {
		return nil, errors.New("question not found")
	}
	return &quiz.Questions[0], nil
}

func (r *quizRepository) UpdateQuestionInQuiz(ctx context.Context, quizID primitive.ObjectID, question *domain.Question) error {
	question.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"questions.$.text":           question.Text,
			"questions.$.options":        question.Options,
			"questions.$.correct_option": question.CorrectOption,
			"questions.$.updated_at":     question.UpdatedAt,
		},
	}
	filter := bson.M{"_id": quizID, "questions._id": question.ID}
	_, err := r.quizzesCollection().UpdateOne(ctx, filter, update)
	return err
}

func (r *quizRepository) DeleteQuestionFromQuiz(ctx context.Context, quizID, questionID primitive.ObjectID) error {
	update := bson.M{"$pull": bson.M{"questions": bson.M{"_id": questionID}}, "$inc": bson.M{"total_questions": -1}}
	_, err := r.quizzesCollection().UpdateOne(ctx, bson.M{"_id": quizID}, update)
	return err
}
