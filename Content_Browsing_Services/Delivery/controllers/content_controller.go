package controllers

import (
	"errors"
	"fmt"
	domain "lawgen/admin-service/Domain"
	usecases "lawgen/admin-service/Usecases"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContentController struct {
	usecase *usecases.ContentUsecase
}

func NewContentController(uc *usecases.ContentUsecase) *ContentController {
	return &ContentController{usecase: uc}
}

// CreateContent handles the ADMIN endpoint for uploading a new PDF.
// POST /api/v1/admin/contents
func (c *ContentController) CreateContent(ctx *gin.Context) {
	// The README specifies multipart/form-data
	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_INPUT", "message": "File is too large (limit 10MB)."})
		return
	}

	file, handler, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "MISSING_FIELD", "message": "Missing 'file' in multipart form."})
		return
	}
	defer file.Close()

	// Extract metadata from form values
	groupName := ctx.Request.FormValue("group_name")
	name := ctx.Request.FormValue("name")
	description := ctx.Request.FormValue("description")
	language := ctx.Request.FormValue("language")
	if name == ""{
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "MISSING_FIELD", "message": "Missing required fields (name)."})
		return
	}
	
	createdContent, err := c.usecase.CreateContent(ctx.Request.Context(), file, handler.Filename, groupName, name, description, language)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "SERVER_ERROR", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Content added.", "id": createdContent.ID, "url": createdContent.URL, "group_id": createdContent.GroupID})
}

// GetAllContent handles the PUBLIC endpoint for listing all available content.
// GET /api/v1/contents
func (c *ContentController) GetAllContent(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	

	response, err := c.usecase.FetchAllGroups(ctx.Request.Context(), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "SERVER_ERROR", "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (c *ContentController) GetContentsByGroupID(ctx *gin.Context) {
	groupID := ctx.Param("groupID")
	page := 1
	limit := 10

	if p := ctx.Query("page"); p != "" {
		fmt.Sscan(p, &page)
	}

	if l := ctx.Query("limit"); l != "" {
		fmt.Sscan(l, &limit)
	}

	contents, err := c.usecase.GetContentsByGroupID(ctx.Request.Context(), groupID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, contents)
}

func (c *ContentController) DeleteContent(ctx *gin.Context) {
    id := ctx.Param("id")

    if err := c.usecase.DeleteContent(ctx, id); err != nil {
        if errors.Is(err, domain.ErrNotFound){
            ctx.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "content deleted successfully"})
}
