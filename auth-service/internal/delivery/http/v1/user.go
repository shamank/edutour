package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/shamank/edutour-backend/auth-service/internal/service"
	"net/http"
)

type userProfileOutput struct {
	UserName   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Avatar     string `json:"avatar"`
	Role       string `json:"role"`
}

type userProfileInput struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Avatar     string `json:"avatar"`
}

func (h *Handler) initUsersRouter(api *gin.RouterGroup) {
	users := api.Group("users")
	{
		users.GET("/:username/profile", h.getUserProfile)
		users.PUT("/:username/profile", h.userIdentity, h.updateUserProfile)

		users.POST("/:username/password", h.userIdentity, h.userChangePassword)
	}
}

// @Summary Get Profile
// @Tags users
// @Description get user profile
// @ModuleID userGetProfile
// @Accept  json
// @Produce  json
// @Param username path string true "username"
// @Success 200 {object} userProfileOutput
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /users/{username}/profile [get]
func (h *Handler) getUserProfile(c *gin.Context) {

	userName := c.Param("username")

	res, err := h.services.Users.GetUserProfile(c.Request.Context(), userName)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, userProfileOutput{
		UserName:   res.UserName,
		FirstName:  res.FirstName,
		LastName:   res.LastName,
		MiddleName: res.MiddleName,
		Avatar:     res.Avatar,
		Role:       res.Role,
	})

}

// @Summary Update Profile
// @Tags users
// @Description update user profile
// @ModuleID userUpdateProfile
// @Accept  json
// @Produce  json
// @Param username path string true "username"
// @Param input body userProfileInput true "update form"
// @Security ApiKeyAuth
// @Success 200 {object} statusResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /users/{username}/profile [put]
func (h *Handler) updateUserProfile(c *gin.Context) {
	var input userProfileInput

	userName := c.Param("username")

	usr, err := h.parseAuthHeader(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	IsAdmin := usr.Role == adminRole

	if usr.userName != userName && !IsAdmin {
		newErrorResponse(c, http.StatusForbidden, "permission denied")
		return
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.services.Users.UpdateUserProfile(c.Request.Context(), userName, service.UserProfileInput{
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		MiddleName: input.MiddleName,
		Avatar:     input.Avatar,
	})
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{"ok"})
}

type userChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=8,max=64"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}

// @Summary Update Password
// @Tags users
// @Description update user password
// @ModuleID userUpdatePassword
// @Accept  json
// @Produce  json
// @Param username path string true "username"
// @Param input body userChangePasswordRequest true "update password form"
// @Security ApiKeyAuth
// @Success 200 {object} statusResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /users/{username}/password [post]
func (h *Handler) userChangePassword(c *gin.Context) {

	var input userChangePasswordRequest

	usr, err := h.parseAuthHeader(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.services.Users.ChangeUserPassword(c.Request.Context(), usr.userID, input.OldPassword, input.NewPassword)

	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: добавить в слой сервиса и репозитория
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})

}
