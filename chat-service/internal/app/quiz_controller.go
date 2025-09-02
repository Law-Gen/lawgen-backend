package app

import (
	"net/http"
	"strconv"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"

	"github.com/gin-gonic/gin"
)

type QuizController struct {
	quizUseCase domain.IQuizUseCase
}

func NewQuizController(quizUseCase domain.IQuizUseCase) *QuizController {
	return &QuizController{quizUseCase: quizUseCase}
}

// --- Public Handler Methods ---

func (c *QuizController) GetCategories(ctx *gin.Context) {
	page, _ := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)

	paginatedCategories, err := c.quizUseCase.ListCategories(ctx.Request.Context(), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, paginatedCategories)
}

func (c *QuizController) GetQuizzesByCategory(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")
	page, _ := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)

	paginatedQuizzes, err := c.quizUseCase.ListQuizzesByCategory(ctx.Request.Context(), categoryID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, paginatedQuizzes)
}

func (c *QuizController) GetQuiz(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	quiz, err := c.quizUseCase.GetQuiz(ctx.Request.Context(), quizID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}
	ctx.JSON(http.StatusOK, quiz)
}

func (c *QuizController) GetQuestionsByQuiz(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	quiz, err := c.quizUseCase.GetQuiz(ctx.Request.Context(), quizID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}
	ctx.JSON(http.StatusOK, quiz.Questions)
}

// type Answers map[string]string

// func (c *QuizController) SubmitQuiz(ctx *gin.Context) {
// 	quizID := ctx.Param("quizId")
// 	var answers Answers
// 	if err := ctx.ShouldBindJSON(&answers); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
// 		return
// 	}

// 	quiz, err := c.quizUseCase.GetQuiz(ctx.Request.Context(), quizID)
// 	if err != nil {
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
// 		return
// 	}

// 	score := 0
// 	for _, question := range quiz.Questions {
// 		if answer, ok := answers[question.ID.Hex()]; ok && answer == question.CorrectOption {
// 			score++
// 		}
// 	}

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"score":          score,
// 		"total_question": len(quiz.Questions),
// 	})
// }

// --- Admin Handler Methods ---

func (c *QuizController) CreateCategory(ctx *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := c.quizUseCase.CreateCategory(ctx.Request.Context(), req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, category)
}

func (c *QuizController) UpdateCategory(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := c.quizUseCase.UpdateCategory(ctx.Request.Context(), categoryID, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, category)
}

func (c *QuizController) DeleteCategory(ctx *gin.Context) {
	categoryID := ctx.Param("categoryId")
	err := c.quizUseCase.DeleteCategory(ctx.Request.Context(), categoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

func (c *QuizController) CreateQuiz(ctx *gin.Context) {
	var req struct {
		CategoryID  string `json:"category_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quiz, err := c.quizUseCase.CreateQuiz(ctx.Request.Context(), req.CategoryID, req.Name, req.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, quiz)
}

func (c *QuizController) UpdateQuiz(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quiz, err := c.quizUseCase.UpdateQuiz(ctx.Request.Context(), quizID, req.Name, req.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, quiz)
}

func (c *QuizController) DeleteQuiz(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	err := c.quizUseCase.DeleteQuiz(ctx.Request.Context(), quizID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

func (c *QuizController) AddQuestion(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	var req struct {
		Text          string            `json:"text" binding:"required"`
		Options       map[string]string `json:"options" binding:"required"`
		CorrectOption string            `json:"correct_option" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quiz, err := c.quizUseCase.AddQuestion(ctx.Request.Context(), quizID, req.Text, req.Options, req.CorrectOption)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, quiz)
}

func (c *QuizController) UpdateQuestion(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	questionID := ctx.Param("questionId")
	var req struct {
		Text          string            `json:"text" binding:"required"`
		Options       map[string]string `json:"options" binding:"required"`
		CorrectOption string            `json:"correct_option" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	question, err := c.quizUseCase.UpdateQuestion(ctx.Request.Context(), quizID, questionID, req.Text, req.Options, req.CorrectOption)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, question)
}

func (c *QuizController) DeleteQuestion(ctx *gin.Context) {
	quizID := ctx.Param("quizId")
	questionID := ctx.Param("questionId")
	err := c.quizUseCase.DeleteQuestion(ctx.Request.Context(), quizID, questionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}
