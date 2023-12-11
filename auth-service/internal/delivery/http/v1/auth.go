package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/shamank/edutour-backend/auth-service/internal/domain"
	"github.com/shamank/edutour-backend/auth-service/internal/service"
	"github.com/shamank/edutour-backend/auth-service/pkg/logger/sl"
	"net/http"
)

type userSignUpInput struct {
	UserName string `json:"username" binding:"required,min=4,max=64"`
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type userSignInInput struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireIn     int    `json:"expire_in"`
}

func (h *Handler) initAuthRouter(api *gin.RouterGroup) {
	auth := api.Group("auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		auth.POST("/confirm", h.confirmUser)
		auth.POST("/reset-password", h.resetPassword)
		auth.POST("/confirm-password", h.confirmResetPassword)

		auth.POST("/refresh", h.userRefresh)

		auth.GET("/me", h.userIdentity, h.userPing)
		auth.GET("/verify", h.userIdentity, h.verifyToken)
	}
}

// @Summary User SignUp
// @Tags auth
// @Description create user account
// @ModuleID authSignUp
// @Accept  json
// @Produce  json
// @Param input body userSignUpInput true "sign up info"
// @Success 201 {string} string "ok"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var input userSignUpInput
	if err := c.BindJSON(&input); err != nil {
		h.logger.Error("cannot bind to json", sl.Err(err))
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Authorization.SignUp(c.Request.Context(), service.UserSignUpInput{
		UserName: input.UserName,
		Email:    input.Email,
		Password: input.Password,
	}); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			newErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{"ok"})
}

// @Summary User SignIn
// @Tags auth
// @Description user sign in
// @ModuleID authSignIn
// @Accept  json
// @Produce  json
// @Param input body userSignInInput true "sign in info"
// @Success 200 {object} tokenResponse
// @Failure 400,401,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {

	var input userSignInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.services.Authorization.SignIn(c.Request.Context(), service.UserSignInInput{
		Login:    input.Login,
		Password: input.Password,
	})

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newErrorResponse(c, http.StatusUnauthorized, err.Error())
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpireIn:     int(res.ExpireIn.Seconds()),
	})
}

type confirmUserRequest struct {
	ConfirmToken string `json:"confirm_token" validate:"required"`
}

// @Summary User Confirm
// @Tags auth
// @Description user confirm email
// @ModuleID authConfirmUser
// @Accept  json
// @Produce  json
// @Param input body confirmUserRequest true "confirm info"
// @Success 200 {object} statusResponse
// @Failure 400,401,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/confirm [post]
func (h *Handler) confirmUser(c *gin.Context) {

	var input confirmUserRequest

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Authorization.ConfirmUser(c.Request.Context(), input.ConfirmToken); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})

	return
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// @Summary Refresh Token
// @Tags auth
// @Description user refresh token
// @ModuleID authRefreshToken
// @Accept  json
// @Produce  json
// @Param input body refreshInput true "refresh token input"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/refresh [post]
func (h *Handler) userRefresh(c *gin.Context) {
	var input refreshInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid body")
		return
	}

	res, err := h.services.Authorization.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpireIn:     int(res.ExpireIn.Seconds()),
	})
}

type resetPasswordRequest struct {
	Email string `json:"email" binding:"email"`
}

// @Summary User reset password
// @Tags auth
// @Description user reset password request
// @ModuleID authResetPassword
// @Accept  json
// @Produce  json
// @Param input body resetPasswordRequest true "reset password input"
// @Success 200 {object} statusResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/reset-password [post]
func (h *Handler) resetPassword(c *gin.Context) {
	var resetPasswordInput resetPasswordRequest

	if err := c.BindJSON(&resetPasswordInput); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.services.Authorization.ResetPassword(c.Request.Context(), resetPasswordInput.Email); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

type confirmPasswordRequest struct {
	ResetToken string `json:"reset_token" binding:"required"`
	Password   string `json:"password" binding:"required,min=8,max=64"`
}

// @Summary User reset password
// @Tags auth
// @Description user reset password confirm
// @ModuleID authConfirmPassword
// @Accept  json
// @Produce  json
// @Param input body confirmPasswordRequest true "reset password input"
// @Success 200 {object} statusResponse
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/confirm-password [post]
func (h *Handler) confirmResetPassword(c *gin.Context) {
	var input confirmPasswordRequest

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Authorization.ConfirmResetPassword(c.Request.Context(), input.ResetToken, input.Password); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

type userPingResponse struct {
	Status   string `json:"status"`
	Username string `json:"username"`
}

// @Summary User check token
// @Tags auth
// @Description user check access token
// @ModuleID authPing
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} userPingResponse
// @Failure 400,401,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/me [get]
func (h *Handler) userPing(c *gin.Context) {
	usrCtx, err := h.parseAuthHeader(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, userPingResponse{
		Status:   "ok",
		Username: usrCtx.userName,
	})
}

// @Summary Verify token for other apps
// @Tags backend
// @Description verify token for other apps
// @ModuleID authVerify
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} statusResponse
// @Failure 400,401,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/verify [get]
func (h *Handler) verifyToken(c *gin.Context) {
	usr, err := h.parseAuthHeader(c)
	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	res, err := h.services.Authorization.GetFullUserInfo(c.Request.Context(), usr.userID)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"id":   res.ID,
		"role": res.Role.ID,
	})

}
