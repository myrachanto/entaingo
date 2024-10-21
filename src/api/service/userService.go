package service

import (
	"github.com/myrachanto/entaingo/src/api/models"
	"github.com/myrachanto/entaingo/src/api/repository"
)

var (
	UserService UserServiceInterface = &userService{}
)

type UserServiceInterface interface {
	Create(transaction *models.TransactionRequest) (*models.UserInfo, error)
	GetTransactions(userId int) (*models.UserInfo, error)
}
type userService struct {
	repo repository.UserrepoInterface
}

func NewUserService(repository repository.UserrepoInterface) UserServiceInterface {
	return &userService{
		repository,
	}
}
func (service *userService) Create(transaction *models.TransactionRequest) (*models.UserInfo, error) {
	return service.repo.Create(transaction)
}
func (service *userService) GetTransactions(userId int) (*models.UserInfo, error) {
	return service.repo.GetTransactions(userId)
}
