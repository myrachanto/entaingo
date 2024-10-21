package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/myrachanto/entaingo/src/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Assuming ValidSources and validSources function are defined in the same package
// Handler to be tested
func TransactionHandler(c *gin.Context) {
	var transaction models.TransactionRequest
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transaction processed successfully"})
}
func TestTransactionHandler(t *testing.T) {
	// Set up a new Gin engine for testing
	router := gin.Default()
	router.POST("/transaction", TransactionHandler)

	// Define test cases
	tests := []struct {
		name           string
		inputJSON      string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid JSON request",
			inputJSON:      `{"state": "win", "amount": 100.5, "transactionId": "tx12345"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Transaction processed successfully"}`,
		},
		{
			name:           "Invalid JSON request - missing amount",
			inputJSON:      `{"state": "completed", "transactionId": "tx12345"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Key: 'TransactionRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag"}`,
		},
		{
			name:           "Invalid JSON request - malformed JSON",
			inputJSON:      `{"state": "completed", "amount": "invalid_amount", "transactionId": "tx12345"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"json: cannot unmarshal string into Go struct field TransactionRequest.amount of type float64"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request with the input JSON
			req, _ := http.NewRequest(http.MethodPost, "/transaction", bytes.NewBufferString(tt.inputJSON))
			req.Header.Set("Content-Type", "application/json")

			// Record the response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check the response body
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
func TestValidSources(t *testing.T) {
	// Test cases
	tests := []struct {
		name       string
		sourceType string
		want       bool
	}{
		{"Valid source - game", "game", true},
		{"Valid source - server", "server", true},
		{"Valid source - payment", "payment", true},
		{"Invalid source - player", "player", false},
		{"Empty source type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validSources(tt.sourceType)
			if got != tt.want {
				t.Errorf("validSources(%s) = %v; want %v", tt.sourceType, got, tt.want)
			}
		})
	}
}

type mockService struct {
	mock.Mock
}

func (m *mockService) Create(transactionReq *models.TransactionRequest) (*models.UserInfo, error) {
	args := m.Called(transactionReq)
	if transaction, ok := args.Get(0).(*models.UserInfo); ok {
		return transaction, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockService) GetTransactions(transactionid int) (*models.UserInfo, error) {
	args := m.Called(transactionid)
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func TestUserController_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		inputBody      interface{}
		sourceType     string
		expectedStatus int
		serviceMock    func(m *mockService)
		expectedError  string
	}{
		{
			name: "invalid json body",
			inputBody: `{
				"transactionID": 123,
				"amount": "string instead of number"
			}`,
			sourceType:     "server",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request",
		},
		{
			name: "invalid source type",
			inputBody: models.TransactionRequest{
				TransactionID: "tx_123",
				Amount:        100,
				State:         "win",
			},
			sourceType:     "web",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid Source-Type",
		},
		{
			name: "valid transaction",
			inputBody: models.TransactionRequest{
				TransactionID: "tx_123",
				Amount:        100,
				State:         "win",
			},
			sourceType: "game",
			serviceMock: func(m *mockService) {
				m.On("Create", mock.Anything).Return(&models.Transaction{
					TransactionID: "tx_123",
					Amount:        100,
					State:         "win",
					UserID:        1,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "transaction already processed",
			inputBody: models.TransactionRequest{
				TransactionID: "tx_123",
				Amount:        100,
				State:         "win",
			},
			sourceType: "server",
			serviceMock: func(m *mockService) {
				m.On("Create", mock.Anything).Return(nil, fmt.Errorf("transaction already processed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "transaction already processed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create mock service
			mockService := new(mockService)
			if test.serviceMock != nil {
				test.serviceMock(mockService)
			}

			// Initialize controller
			controller := userController{
				service: mockService,
			}
			// controller := NewUserController(mockService)

			// Set up router and recorder
			router := gin.Default()
			router.POST("/transaction", controller.Create)

			// Prepare request body
			bodyBytes, _ := json.Marshal(test.inputBody)
			req, _ := http.NewRequest(http.MethodPost, "/transaction", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Source-Type", test.sourceType)
			req.Header.Set("Content-Type", "application/json")

			// Record the response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, test.expectedStatus, w.Code)

			if test.expectedError != "" {
				var response map[string]string
				_ = json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], test.expectedError)
			}
		})
	}
}
