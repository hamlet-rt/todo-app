package handler

import (
	"net/http"
	"todo"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @Sumary SignUp
// @Tags auth
// @Description create account
// @ID create-account
// @Accepte json
// @Produce json
// @Param input body todo.User true "account info"
// @Success 200 {integer} integer 1
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var input todo.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Authorization.CreateUser(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type signInInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @Sumary SignIn
// @Tags auth
// @Description login
// @ID login
// @Accepte json
// @Produce json
// @Param input body signInInput true "credentials"
// @Success 200 {integer} integer 1
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var input signInInput

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.services.Authorization.GetUser(input.Username, input.Password)
    if err != nil {
        logrus.Errorf("failed to get user: %v", err)
        newErrorResponse(c, http.StatusUnauthorized, "invalid username or password")
        return
    }

	token, refreshToken, err := h.services.Authorization.GenerateTokens(user.Id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("refreshToken", refreshToken, 3600*24*30, "/auth", "", false, true)

	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}

// @Sumary Refresh
// @Tags auth
// @Description refreshToken
// @ID refresh-token
// @Accepte json
// @Produce json
// @Param refreshToken header string true "Refresh token in the cookie"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errorResponse "Refresh token not found"
// @Failure 500 {object} errorResponse "Failed to refresh tokens"
// @Router /auth/refresh [post]
func (h *Handler) refresh(c *gin.Context) {
    cookie, err := c.Cookie("refreshToken")
    if err != nil {
        logrus.Errorf("failed to get refresh-token cookie: %v", err)
        newErrorResponse(c, http.StatusUnauthorized, "refresh token not found")
        return
    }

    logrus.Infof("Refresh token from cookie: %s", cookie)

    refreshToken := todo.RefreshToken{
        Token: cookie,
    }

    accessToken, newRefreshToken, err := h.services.Authorization.RefreshTokens(refreshToken)
    if err != nil {
        logrus.Errorf("failed to refresh tokens: %v", err)
        newErrorResponse(c, http.StatusInternalServerError, "failed to refresh tokens")
        return
    }

	c.SetCookie("refreshToken", newRefreshToken, 3600*24*30, "/auth", "", false, true)

    c.JSON(http.StatusOK, map[string]interface{}{
        "accessToken":  accessToken,
    })
}
