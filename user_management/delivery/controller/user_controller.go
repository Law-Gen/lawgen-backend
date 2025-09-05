package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"
	"user_management/domain"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userUsecase domain.UserUsecase
}

func NewUserController(uuc domain.UserUsecase) *UserController {
	return &UserController{
		userUsecase: uuc,
	}
}

type PaginationDTO struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

type AllusersResponse struct {
	Allusers       []UserDTO
	PaginationData PaginationDTO
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func ConvertDTOSlicetoDomian(users []domain.User) []UserDTO {
	domainUsers := make([]UserDTO, len(users))
	for i, user := range users {
		domainUsers[i] = *ConvertToUserDTO(&user)
	}
	return domainUsers
}

func (uc *UserController) ChangeUserRole(c *gin.Context, roleChange func(context.Context, string, string) error, successMessage string) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userid, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req EmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()
	if err := roleChange(ctx, userid, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": successMessage})
}

func (uc *UserController) HandlePromote(c *gin.Context) {
	uc.ChangeUserRole(c, uc.userUsecase.Promote, "user promoted successfully")
}

func (uc *UserController) HandleDemote(c *gin.Context) {
	uc.ChangeUserRole(c, uc.userUsecase.Demote, "user demoted successfully")
}

func (uc *UserController) HandleUpdateUser(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	gender := c.PostForm("gender")
	birthDate := c.PostForm("birth_date")

	// Validate and parse birthDate format (YYYY-MM-DD)
	var birth_date time.Time
	if birthDate != "" {
		parsedDate, err := time.Parse("2006-01-02", birthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birth_date format. Use YYYY-MM-DD."})
			return
		}
		birth_date = parsedDate
	}

	languagePreference := c.PostForm("language_preference")

	file, err := c.FormFile("profile_picture")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileReader.Close()

	ctx := c.Request.Context()
	if err := uc.userUsecase.ProfileUpdate(ctx, userID, gender, birth_date, languagePreference, fileReader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (uc *UserController) HandleGetAllUsers(c *gin.Context) {
	var defaultPage = 1
	var defaultLimit = 10

	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	if err != nil || page < 1 {
		page = defaultPage
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 {
		limit = defaultLimit
	}

	allusers, total, err := uc.userUsecase.GetAllUsers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	all := ConvertDTOSlicetoDomian(allusers)
	pagination := PaginationDTO{
		Total: total,
		Page:  page,
		Limit: limit,
	}
	res := AllusersResponse{
		Allusers:       all,
		PaginationData: pagination,
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (uc *UserController) HandleGetUserByID(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx := c.Request.Context()
	user, err := uc.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ConvertToUserDTO(user)})
}

func (uc *UserController) HandleChangePassword(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := uc.userUsecase.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
