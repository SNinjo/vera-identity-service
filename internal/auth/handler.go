package auth

import (
	"net/http"
	"strconv"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/config"
	"vera-identity-service/internal/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config      *config.Config
	authService Service
	userService user.Service
}

func NewHandler(config *config.Config, authService Service, userService user.Service) *Handler {
	return &Handler{config: config, authService: authService, userService: userService}
}

func (h *Handler) Login(c *gin.Context) {
	c.Redirect(http.StatusFound, h.authService.GetOAuthLoginURL())
}

func (h *Handler) Callback(c *gin.Context) {
	code := c.Query("code")

	idTokenClaims, err := h.authService.GetOAuthIDTokenClaims(code)
	if err != nil {
		c.Error(err)
		return
	}

	user, err := h.userService.GetUserByEmail(idTokenClaims.Email)
	if err != nil {
		c.Error(err)
		return
	}
	if user == nil {
		c.Error(apperror.New(apperror.CodeUserNotAuthorized, "user not authorized | email: "+idTokenClaims.Email))
		return
	}

	if err = h.userService.RecordUserLogin(user.ID, idTokenClaims.Name, idTokenClaims.Picture, idTokenClaims.Subject); err != nil {
		c.Error(err)
		return
	}

	accessToken, err := h.authService.NewAccessToken(user.ID, idTokenClaims.Name, idTokenClaims.Email, idTokenClaims.Picture)
	if err != nil {
		c.Error(err)
		return
	}
	refreshToken, err := h.authService.NewRefreshToken(user.ID)
	if err != nil {
		c.Error(err)
		return
	}

	c.SetCookie("refresh_token", refreshToken, int(h.config.RefreshTokenTTL.Seconds()), "/", "", true, true)
	c.Redirect(http.StatusFound, h.config.SiteURL+"?access_token="+accessToken)
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")
	claims, err := h.authService.ParseRefreshToken(refreshToken)
	if err != nil {
		c.Error(apperror.New(apperror.CodeInvalidRefreshToken, "invalid refresh token | refresh token: "+refreshToken))
		return
	}

	userID, _ := strconv.Atoi(claims.Subject)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.Error(err)
		return
	} else if user == nil {
		c.Error(apperror.New(apperror.CodeUserNotAuthorized, "user not authorized | id: "+strconv.Itoa(userID)))
		return
	}

	accessToken, err := h.authService.NewAccessToken(user.ID, *user.Name, user.Email, *user.Picture)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{AccessToken: accessToken})
}

func (h *Handler) Verify(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
