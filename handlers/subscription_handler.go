package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"subscription_service/database"
	"subscription_service/models"
)

type SubscriptionHandler struct {
	db     *database.DB
	logger *zap.Logger
}

func NewSubscriptionHandler(db *database.DB, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		db:     db,
		logger: logger,
	}
}

// @Summary Создать подписку
// @Description Создает новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Данные подписки"
// @Success 201 {object} models.SubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Ошибка валидации запроса", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.logger.Warn("Неверный формат UUID пользователя", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат user_id"})
		return
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		h.logger.Warn("Неверный формат даты начала", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат start_date. Ожидается MM-YYYY"})
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		ed, err := parseMonthYear(req.EndDate)
		if err != nil {
			h.logger.Warn("Неверный формат даты окончания", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат end_date. Ожидается MM-YYYY"})
			return
		}
		endDate = &ed
	}

	id, err := h.db.CreateSubscription(req.ServiceName, req.Price, userID, &startDate, endDate)
	if err != nil {
		h.logger.Error("Ошибка создания подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка создания подписки"})
		return
	}

	sub, err := h.db.GetSubscription(*id)
	if err != nil {
		h.logger.Error("Ошибка получения созданной подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения подписки"})
		return
	}

	c.JSON(http.StatusCreated, toSubscriptionResponse(sub))
}

// @Summary Получить подписку
// @Description Получает подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} models.SubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn("Неверный формат UUID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID"})
		return
	}

	sub, err := h.db.GetSubscription(id)
	if err != nil {
		if err.Error() == "Подписка не найдена" {
			c.JSON(http.StatusNotFound, gin.H{"error": "подписка не найдена"})
			return
		}
		h.logger.Error("Ошибка получения подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения подписки"})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(sub))
}

// @Summary Обновить подписку
// @Description Обновляет данные подписки
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param subscription body models.UpdateSubscriptionRequest true "Данные для обновления"
// @Success 200 {object} models.SubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn("Неверный формат UUID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Ошибка валидации запроса", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startDate *time.Time
	if req.StartDate != nil {
		sd, err := parseMonthYear(*req.StartDate)
		if err != nil {
			h.logger.Warn("Неверный формат даты начала", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат start_date. Ожидается MM-YYYY"})
			return
		}
		startDate = &sd
	}

	var endDate *time.Time
	if req.EndDate != nil {
		ed, err := parseMonthYear(*req.EndDate)
		if err != nil {
			h.logger.Warn("Неверный формат даты окончания", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат end_date. Ожидается MM-YYYY"})
			return
		}
		endDate = &ed
	}

	err = h.db.UpdateSubscription(id, req.ServiceName, req.Price, startDate, endDate)
	if err != nil {
		if err.Error() == "Подписка не найдена" {
			c.JSON(http.StatusNotFound, gin.H{"error": "подписка не найдена"})
			return
		}
		h.logger.Error("Ошибка обновления подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка обновления подписки"})
		return
	}

	sub, err := h.db.GetSubscription(id)
	if err != nil {
		h.logger.Error("Ошибка получения обновленной подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения подписки"})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(sub))
}

// @Summary Удалить подписку
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.logger.Warn("Неверный формат UUID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID"})
		return
	}

	err = h.db.DeleteSubscription(id)
	if err != nil {
		if err.Error() == "Подписка не найдена" {
			c.JSON(http.StatusNotFound, gin.H{"error": "подписка не найдена"})
			return
		}
		h.logger.Error("Ошибка удаления подписки", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка удаления подписки"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Список подписок
// @Description Получает список всех подписок с пагинацией
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Лимит записей" default(10)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.SubscriptionResponse
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	subscriptions, err := h.db.ListSubscriptions(limit, offset)
	if err != nil {
		h.logger.Error("Ошибка получения списка подписок", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения списка подписок"})
		return
	}

	response := make([]models.SubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		response[i] = toSubscriptionResponse(&sub)
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Подсчет суммарной стоимости
// @Description Вычисляет суммарную стоимость всех подписок за выбранный период с фильтрацией
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса"
// @Param start_period query string true "Начало периода (MM-YYYY)"
// @Param end_period query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	var req models.TotalCostRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("Ошибка валидации запроса", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startPeriod, err := parseMonthYear(req.StartPeriod)
	if err != nil {
		h.logger.Warn("Неверный формат начального периода", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат start_period. Ожидается MM-YYYY"})
		return
	}

	endPeriod, err := parseMonthYear(req.EndPeriod)
	if err != nil {
		h.logger.Warn("Неверный формат конечного периода", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат end_period. Ожидается MM-YYYY"})
		return
	}

	endPeriod = time.Date(endPeriod.Year(), endPeriod.Month()+1, 0, 23, 59, 59, 999999999, time.UTC)

	var userID *uuid.UUID
	if req.UserID != nil {
		uid, err := uuid.Parse(*req.UserID)
		if err != nil {
			h.logger.Warn("Неверный формат UUID пользователя", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат user_id"})
			return
		}
		userID = &uid
	}

	totalCost, err := h.db.CalculateTotalCost(userID, req.ServiceName, startPeriod, endPeriod)
	if err != nil {
		h.logger.Error("Ошибка вычисления суммарной стоимости", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка вычисления суммарной стоимости"})
		return
	}

	c.JSON(http.StatusOK, models.TotalCostResponse{TotalCost: totalCost})
}

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func toSubscriptionResponse(sub *models.Subscription) models.SubscriptionResponse {
	response := models.SubscriptionResponse{
		ID:          sub.ID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   sub.StartDate.Format(time.RFC3339),
		CreatedAt:   sub.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   sub.UpdatedAt.Format(time.RFC3339),
	}

	if sub.EndDate != nil {
		endDateStr := sub.EndDate.Format(time.RFC3339)
		response.EndDate = &endDateStr
	}

	return response
}

