package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/satrunjis/user-service/internal/domain"
	"github.com/satrunjis/user-service/internal/service"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	userService service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: *userService,
		logger:      logger,
	}
}

// GetUsers godoc
// @Summary      Поиск и фильтрация пользователей
// @Description  Полнотекстовый поиск с фильтрацией и сортировкой
// @Tags         users
// @Accept       json
// @Produce      json
// @Param   filters query domain.UserFilter false "Фильтры"
// @Success      200         {object}  UserListResponse
// @Failure      500         {object}  ErrorResponse
// @Router       /api/v1/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	filters := domain.UserFilter{
		Search:     strPtr(c.Query("q")),
		DateFrom:   parseTimePtr(c.Query("date_from")),
		DateTo:     parseTimePtr(c.Query("date_to")),
		Distance:   strPtr(c.Query("radius")),
		Lat:        parsefloatPtr(c.Query("lat")),
		Lon:        parsefloatPtr(c.Query("lon")),
		SocialType: strPtr(c.Query("social_net")),
		SortBy:     strPtr(c.Query("sort_by")),
		SortOrder:  strPtr(c.Query("sort_order")),
		Page:       parseIntPtr(c.Query("page")),
		Size:       parseIntPtr(c.Query("size")),
	}

	users, err := h.userService.SearchUsers(c.Request.Context(), &filters)
	if err != nil {
		h.logger.Error("Failed to get users", "err", err)
		c.Error(err)
		return
	}

	userVals := make([]domain.User, len(users))
	for i, u := range users {
		if u != nil {
			userVals[i] = *u
		}
	}
	//h.logger.Debug("Filtered users", "filters", filters.String())

	c.JSON(http.StatusOK, UserListResponse{Users: userVals, Total: len(userVals)})
}

// CreateUser godoc
// @Summary      Создать пользователя
// @Description  Создать новую запись о пользователе
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body  domain.User  false  "Данные пользователя"
// @Success      201   {object}  UserID
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	const op = "UserHandler.CreateUser"
	log := h.logger.With("operation", op)
	start := time.Now()
	ctx := c.Request.Context()

	log.DebugContext(ctx, "start creating user")

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.ErrorContext(ctx, "failed to bind JSON", "error", err)
		c.Error(&service.ServiceError{
			Code:    service.ErrCodeInvalidInput,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}

	log.DebugContext(ctx, "JSON bind successful", "user_id", user.ID)

	if err := h.userService.CreateUser(ctx, &user); err != nil {
		log.ErrorContext(ctx, "failed to create user", "error", err, "user_id", user.ID)
		c.Error(err)
		return
	}

	log.InfoContext(ctx, "user created successfully",
		"user_id", user.ID,
		"duration", time.Since(start))

	c.JSON(http.StatusCreated, UserID{UserID: *user.ID})
}

// GetUser godoc
// @Summary      Получить пользователя
// @Description  Получить пользователя по ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  domain.User
// @Failure      404  {object}  ErrorResponse
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userService.GetUserByID(c.Request.Context(), &id)
	if err != nil {
		h.logger.Error("Failed to get user", "err", err)
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Обновить пользователя
// @Description  Полное обновление данных пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string        true  "User ID"
// @Param        user body      UserWithoutID  false  "Обновлённые данные"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	const op = "UserHandler.UpdateUser"
	log := h.logger.With("operation", op)
	ctx := c.Request.Context()
	start := time.Now()

	log.DebugContext(ctx, "start user update")

	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.ErrorContext(ctx, "failed to bind JSON", "error", err)
		c.Error(&service.ServiceError{
			Code:    service.ErrCodeInvalidInput,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}

	user.ID = &id

	log.DebugContext(ctx, "JSON bind successful", "user_id", *user.ID)

	if err := h.userService.Replace(ctx, &user); err != nil {
		log.ErrorContext(ctx, "failed to replace user", "error", err, "user_id", *user.ID)
		c.Error(err)
		return
	}

	log.InfoContext(ctx, "user updated successfully",
		"user_id", *user.ID,
		"duration", time.Since(start))

	c.JSON(http.StatusOK, user)
}

// UpdateUserPartial godoc
// @Summary      Частично обновить пользователя
// @Description  Частичное обновление данных пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Param        patch body     domain.User  true  "Поля для обновления"
// @Success      200  {object}  UserWithoutID
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/users/{id} [patch]
func (h *UserHandler) UpdateUserPartial(c *gin.Context) {
	id := c.Param("id")
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Error("Failed to bind JSON", "err", err)
		c.Error(&service.ServiceError{
			Code:    service.ErrCodeInvalidInput,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}
	user.ID = &id
	if err := h.userService.UpdatePartial(c.Request.Context(), &user); err != nil {
		h.logger.Error("Failed to update user", "err", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      Удалить пользователя
// @Description  Удалить запись о пользователе
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := h.userService.DeleteUser(c.Request.Context(), &id)
	if err != nil {
		h.logger.Error("Failed to delete user", "err", err)
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetUserMap godoc
// @Summary      Получить карту пользователя
// @Description  Получить PNG-карту местоположения пользователя
// @Tags         users
// @Accept       json
// @Produce      image/png
// @Param        id   path      string  true  "User ID"
// @Success      200  {file}   body  "PNG изображение"
// @Failure      404  {object}   ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/users/{id}/map [get]
func (h *UserHandler) GetUserMap(c *gin.Context) {
	id := c.Param("id")
	zoom := 13
	maptile, err := h.userService.GetMapTile(c.Request.Context(), &id, zoom)
	if err != nil {
		h.logger.Error("Failed to get map user", "err", err)
		c.Error(err)
		return
	}
	if maptile == nil {
		c.Error(&service.ServiceError{
			Code:    service.ErrCodeInternal,
			Message: "Empty map data",
		})
		return
	}
	c.Data(http.StatusOK, "image/png", *maptile)
}

type UserListResponse struct {
	Users []domain.User `json:"users"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
}

type UserID struct {
	UserID string `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &i
}

func parsefloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

type UserWithoutID struct {
	Login       *string    `form:"login" json:"login,omitempty" example:"john_doe" swagger:"description='Логин пользователя'"`
	Username    *string    `form:"username" json:"username,omitempty" example:"John Doe" swagger:"description='Имя пользователя'"`
	Password    *string    `form:"password" json:"password,omitempty" example:"secret123" swagger:"description='Пароль пользователя '"`
	Description *string    `form:"description" json:"description,omitempty" example:"Программист из Санкт-Петербурга" swagger:"description='Описание пользователя'"`
	Comment     *string    `form:"comment" json:"comment,omitempty" example:"Важный клиент" swagger:"description='Комментарии о пользователе (заметка админа)'"`
	RegDate     *time.Time `form:"reg_date" json:"reg_date,omitempty" example:"2023-01-15T12:34:56Z" swagger:"description='Дата регистрации'"`
	Location    *domain.Location  `form:"location" json:"location,omitempty" swagger:"description='Геолокация пользователя'"`
	SocialNet   *string    `form:"social_net" json:"social_net,omitempty" example:"MAX" swagger:"description='Название соц. сети, строго определенное'"`
}
