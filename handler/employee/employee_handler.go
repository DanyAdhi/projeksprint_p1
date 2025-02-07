package employeeHandler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/levensspel/go-gin-template/dto"
	"github.com/levensspel/go-gin-template/helper"
	"github.com/levensspel/go-gin-template/logger"
	"github.com/levensspel/go-gin-template/middleware"
	service "github.com/levensspel/go-gin-template/service/employee"
	"github.com/levensspel/go-gin-template/validation"
	"github.com/samber/do/v2"
)

type EmployeeHandler interface {
	Create(ctx *gin.Context)
	GetAll(ctx *gin.Context)
}

type handler struct {
	service service.EmployeeService
	logger  logger.Logger
}

func NewEmployeeHandler(service service.EmployeeService, logger logger.Logger) EmployeeHandler {
	return &handler{service: service, logger: logger}
}

func NewEmployeeHandlerInject(i do.Injector) (EmployeeHandler, error) {
	_service := do.MustInvoke[service.EmployeeService](i)
	_logger := do.MustInvoke[logger.LogHandler](i)
	return NewEmployeeHandler(_service, &_logger), nil
}

// Create a new employee
// @Tags employee
// @Summary Create a new employee
// @Description Create a new employee
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param data body dto.EmployeePayload true "data"
// @Success 201 {object} helper.Response{data=helper.Response} "Created"
// @Failure 400 {object} helper.Response{errors=helper.ErrorResponse} "Bad Request"
// @Failure 401 {object} helper.Response{errors=helper.ErrorResponse} "Unauthorized"
// @Failure 409 {object} helper.Response{errors=helper.ErrorResponse} "Conflict"
// @Failure 500 {object} helper.Response{errors=helper.ErrorResponse} "Server Error"
// @Router /v1/employee [POST]
func (h *handler) Create(ctx *gin.Context) {
	defer helper.FallbackResponse(ctx)

	managerID, err := middleware.GetIdUserFromContext(ctx)
	if err != nil {
		h.logger.Warn(err.Error(), helper.EmployeeHandlerCreate)
		ctx.JSON(helper.GetErrorStatusCode(helper.ErrUnauthorized), helper.NewResponse(nil, err))
		return
	}

	input := new(dto.EmployeePayload)

	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.logger.Warn(err.Error(), helper.EmployeeHandlerCreate, input)
		ctx.JSON(helper.GetErrorStatusCode(helper.ErrBadRequest), helper.NewResponse(nil, err))
		return
	}

	err = validation.ValidateEmployeeCreate(input)
	if err != nil {
		h.logger.Warn(err.Error(), helper.EmployeeHandlerCreate, input)
		ctx.JSON(helper.GetErrorStatusCode(helper.ErrBadRequest), helper.NewResponse(nil, err))
		return
	}

	err = h.service.Create(ctx, *input, managerID)
	if err != nil {
		h.logger.Error(err.Error(), helper.EmployeeHandlerCreate)
		ctx.JSON(
			helper.GetErrorStatusCode(err),
			helper.NewResponse(
				nil,
				errors.New((helper.GetErrorMessage(err)))),
		)
		return
	}

	ctx.JSON(http.StatusOK, input)
	return
}

// Get employee
// @Tags employee
// @Summary Get employee
// @Description Get employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer + user token"
// @Param data body dto.GetEmployeesRequest true "data"
// @Success 200 {object} helper.Response{data=helper.Response} "OK"
// @Failure 400 {object} helper.Response{errors=helper.ErrorResponse} "Bad Request"
// @Failure 401 {object} helper.Response{errors=helper.ErrorResponse} "Unauthorization"
// @Router /v1/employee [GET]
func (h handler) GetAll(ctx *gin.Context) {
	defer helper.FallbackResponse(ctx)

	input := new(dto.GetEmployeesRequest)

	setGetEmployeeRequest(ctx, input)

	err := validation.ValidateEmployeeGet(input)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			helper.NewResponse(
				helper.ErrorResponse{
					Code:    helper.GetErrorStatusCode(err),
					Message: err.Error(),
				},
				err,
			),
		)
		return
	}

	response, err := h.service.GetAll(ctx, *input)

	if err != nil {
		ctx.JSON(
			helper.GetErrorStatusCode(err),
			helper.NewResponse(
				nil,
				errors.New((helper.GetErrorMessage(err)))),
		)
		return
	}
	ctx.JSON(http.StatusOK, helper.NewResponse(response, nil))
}

func setGetEmployeeRequest(ctx *gin.Context, input *dto.GetEmployeesRequest) {
	managerId, err := middleware.GetIdUserFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, helper.NewResponse(nil, err))
		return
	}
	input.ManagerID = managerId

	gender := ctx.Request.URL.Query().Get("gender")
	input.Gender = strings.ToLower(gender)

	idNumber := ctx.Request.URL.Query().Get("identityNumber")
	input.IdentityNumber = strings.ToLower(idNumber)

	name := ctx.Request.URL.Query().Get("name")
	input.Name = strings.ToLower(name)

	departmentId := ctx.Request.URL.Query().Get("departmentId")
	input.DepartmentID = strings.ToLower(departmentId)

	limitParam := ctx.Request.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 0 {
		input.Limit = dto.DefaultLimit
	} else {
		input.Limit = limit
	}

	offsetParam := ctx.Request.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		input.Offset = dto.DefaultOffset
	} else {
		input.Offset = offset
	}
}
