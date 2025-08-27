package controllers

import (
	domain "lawgen/admin-service/Domain"
	usecases "lawgen/admin-service/Usecases"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LegalEntityController struct {
	usecase *usecases.LegalEntityUsecase
}

func NewLegalEntityController (u *usecases.LegalEntityUsecase) *LegalEntityController {
	return &LegalEntityController{ usecase: u }
}

func (ctrl *LegalEntityController) CreateLegalEntity(c *gin.Context) {
	var entity *domain.LegalEntity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid request: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	result, err := ctrl.usecase.CreateLegalEntity(ctx, entity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (ctrl *LegalEntityController) FetchLegalEntityById(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	entity, err := ctrl.usecase.FetchLegalEntityById(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (ctrl *LegalEntityController) FetchAllLegalEntities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page","1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit","10"))
	search := c.DefaultQuery("search", "")
	ctx := c.Request.Context()

	results, err := ctrl.usecase.FetchAllLegalEntities(ctx, page, limit,search)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (ctrl *LegalEntityController) UpdateLegalEntity(c *gin.Context) {
	id := c.Param("id")
	var entity domain.LegalEntity
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"invalid request": err.Error()})
		return
	}

	ctx := c.Request.Context()

	if err := ctrl.usecase.UpdateLegalEntity(ctx, id, &entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message":"Updated Successfully"})
}

func (ctrl *LegalEntityController) DeleteLegalEntity(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if err := ctrl.usecase.DeleteLegalEntity(ctx, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message":"Deleted Successfully"})
}