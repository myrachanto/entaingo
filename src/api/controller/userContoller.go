package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/myrachanto/entaingo/src/api/models"
	"github.com/myrachanto/entaingo/src/api/service"
)

// UserController ...
var (
	UserController UserControllerInterface = &userController{}
	ValidSources                           = []string{"game", "server", "payment"}
)

type UserControllerInterface interface {
	Create(c *gin.Context)
	GetTransactions(c *gin.Context)
}

type userController struct {
	service service.UserServiceInterface
}

func NewUserController(ser service.UserServiceInterface) UserControllerInterface {
	return &userController{
		ser,
	}
}

// ///////controllers/////////////////

// Create godoc
// @Summary Create a transaction
// @Description Create a new transaction item
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.TransactionRequest true "Transaction Request"
// @Success 201 {object} models.UserInfo "Transaction created"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /transaction [post]
func (controller userController) Create(c *gin.Context) {
	transaction := &models.TransactionRequest{}
	// Parse the request body
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// check for source validity
	sourceType := c.GetHeader("Source-Type")
	ok := validSources(sourceType)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Source-Type"})
		return
	}
	transaction.SourceType = sourceType

	res, err := controller.service.Create(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

// Get godoc
// @Summary Get transaction details for a user
// @Description Retrieve transactions for a specific user by ID
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.UserInfo
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /transaction/{id} [get]
func (controller userController) GetTransactions(c *gin.Context) {
	id := c.Param("id")
	Id, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "failed to parse the id"})
		return
	}
	userInfo, err := controller.service.GetTransactions(int(Id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"userInfo": userInfo})
}

func validSources(sourceType string) bool {

	isValidSource := false
	for _, source := range ValidSources {
		if sourceType == source {
			isValidSource = true
			break
		}
	}
	return isValidSource

}
