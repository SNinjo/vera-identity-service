package user

import (
	"net/http"
	"strconv"
	"time"
	"vera-identity-service/internal/apperror"

	"github.com/gin-gonic/gin"
)

func loginHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, getOAuthLoginURL())
}

func callbackHandler(c *gin.Context) {
	code := c.Query("code")

	resp, err := handleOAuthCallback(code)
	if err != nil {
		c.Error(err)
		return
	}
	user, err := getUserByEmail(resp.Email)
	if err != nil {
		c.Error(err)
		return
	} else if user == nil {
		c.Error(apperror.New(apperror.CodeUserNotFound, "user not found | email: "+resp.Email))
		return
	}

	now := time.Now()
	newUser, err := updateUser(user.ID, &User{
		Email:        resp.Email,
		Name:         &resp.Name,
		Picture:      resp.Picture,
		LastLoginSub: &resp.Sub,
		LastLoginAt:  &now,
	})
	if err != nil {
		c.Error(err)
		return
	}

	accessToken, err := newJWT(
		&User{ID: newUser.ID, Name: newUser.Name, Email: newUser.Email, Picture: newUser.Picture},
		authConfig.AccessTokenSecret,
		authConfig.AccessTokenTTL,
	)
	if err != nil {
		c.Error(err)
		return
	}
	refreshToken, err := newJWT(
		&User{ID: newUser.ID},
		authConfig.RefreshTokenSecret,
		authConfig.RefreshTokenTTL,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.SetCookie("refresh_token", refreshToken, int(authConfig.RefreshTokenTTL.Seconds()), "/", "", true, true)
	c.Redirect(http.StatusFound, authConfig.FrontendURL+"?access_token="+accessToken)
}

func refreshHandler(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")
	claims, err := parseJWT(refreshToken, authConfig.RefreshTokenSecret)
	if err != nil {
		c.Error(apperror.New(apperror.CodeInvalidRefreshToken, "invalid refresh token | refresh token: "+refreshToken))
		return
	}

	userID, _ := strconv.Atoi(claims.Subject)
	user, err := getUserByID(userID)
	if err != nil {
		c.Error(err)
		return
	} else if user == nil {
		c.Error(apperror.New(apperror.CodeUserNotFound, "user not found | id: "+strconv.Itoa(userID)))
		return
	}

	accessToken, err := newJWT(
		&User{ID: user.ID, Email: user.Email, Name: user.Name, Picture: user.Picture},
		authConfig.AccessTokenSecret,
		authConfig.AccessTokenTTL,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, tokenResponse{AccessToken: accessToken})
}

func verifyHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func listUsersHandler(c *gin.Context) {
	users, err := getUsers()
	if err != nil {
		c.Error(err)
		return
	}

	userResponses := make([]userResponse, len(users))
	for i, user := range users {
		userResponses[i] = *newUserResponse(&user)
	}

	c.JSON(http.StatusOK, userResponses)
}

func createUserHandler(c *gin.Context) {
	var req requestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body | " + err.Error()})
		return
	}

	existingUser, err := getUserByEmail(req.Email)
	if err != nil {
		c.Error(err)
		return
	} else if existingUser != nil {
		c.Error(apperror.New(apperror.CodeUserAlreadyExists, "user already exists | email: "+req.Email))
		return
	}

	newUser := &User{Email: req.Email}
	err = createUser(newUser)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, newUserResponse(newUser))
}

func updateUserHandler(c *gin.Context) {
	var uri requestUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}
	var body requestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body | " + err.Error()})
		return
	}

	existingUser, err := getUserByEmail(body.Email)
	if err != nil {
		c.Error(err)
		return
	}
	if existingUser != nil && existingUser.ID != uri.ID {
		c.Error(apperror.New(apperror.CodeUserAlreadyExists, "Email already taken by another user"))
		return
	}

	newUser, err := updateUser(uri.ID, &User{Email: body.Email})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, newUserResponse(newUser))
}

func deleteUserHandler(c *gin.Context) {
	var req requestUri
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}

	err := deleteUser(req.ID)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
