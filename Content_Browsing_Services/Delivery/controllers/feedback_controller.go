package controllers

import (
	Domain "lawgen/admin-service/Domain"
	Usecases "lawgen/admin-service/Usecases"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeedbackController struct {
    usecase *Usecases.FeedbackUsecase
}

func NewFeedbackController(uc *Usecases.FeedbackUsecase) *FeedbackController {
    return &FeedbackController{usecase: uc}
}

func (fc *FeedbackController) CreateFeedback(c *gin.Context) {
    var feedback Domain.Feedback
    if err := c.ShouldBindJSON(&feedback); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := fc.usecase.CreateFeedback(&feedback); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, feedback)
}

func (fc *FeedbackController) GetFeedbackByID(c *gin.Context) {
    id := c.Param("id")
    feedback, err := fc.usecase.GetFeedbackByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
        return
    }
    c.JSON(http.StatusOK, feedback)
}

func (fc *FeedbackController) ListFeedbacks(c *gin.Context) {
    feedbacks, err := fc.usecase.ListFeedbacks()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, feedbacks)
}