package handler

import (
	"backend-viblo-trending/custom_error"
	"backend-viblo-trending/log"
	"backend-viblo-trending/model"
	"backend-viblo-trending/model/requests"
	"backend-viblo-trending/repository"
	security "backend-viblo-trending/security"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/google/uuid"
	"github.com/labstack/echo"
)

type UserHandler struct {
	UserRepo repository.UserRepo
}

func (u *UserHandler) SignUp(c echo.Context) error {
	req := requests.RequestSignUp{}
	if err := c.Bind(&req); err != nil {
		log.Error(err.Error())
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(req); err != nil {
		log.Error(err.Error())
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	hash := security.HashAndSalt([]byte(req.Password))
	role := model.MEMBER.String()

	userId, err := uuid.NewUUID()
	if err != nil {
		log.Error(err.Error())
		return c.JSON(http.StatusForbidden, model.Response{
			StatusCode: http.StatusForbidden,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	user := model.User{
		UserId:   userId.String(),
		FullName: req.FullName,
		Email:    req.Email,
		Password: hash,
		Role:     role,
	}

	user, err = u.UserRepo.SaveUser(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusConflict, model.Response{
			StatusCode: http.StatusConflict,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Đăng ký thành công",
		Data:       user,
	})
}

func (u *UserHandler) SignIn(c echo.Context) error {
	req := requests.RequestSignIn{}
	if err := c.Bind(&req); err != nil {
		log.Error(err.Error())
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	if err := c.Validate(req); err != nil {
		log.Error(err.Error())
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	user, err := u.UserRepo.CheckLogin(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	// check password
	isTheSame := security.ComparePasswords(user.Password, []byte(req.Password))
	if !isTheSame {
		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    "Đăng nhập thất bại",
			Data:       nil,
		})
	}

	// gen access token
	accessToken, err := security.CreateAccessToken(user)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, model.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
			Data:       nil,
		})
	}
	user.AccessToken = accessToken

	// create the access cookie for client(browser)
	accessTokenCookie := &http.Cookie{
		Name:     "AccessToken",
		Value:    accessToken,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	}
	c.SetCookie(accessTokenCookie)

	// gen refresh token
	refreshToken, err := security.CreateRefeshToken(user)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusInternalServerError, model.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
			Data:       nil,
		})
	}
	user.RefreshToken = refreshToken

	// create the refresh cookie for client(browser)
	refreshTokenCookie := &http.Cookie{
		Name:     "RefreshToken",
		Value:    refreshToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
	c.SetCookie(refreshTokenCookie)

	// After the cookie is created, the client(browser) will send in the cookie
	// for every request. Our server side program will unpack the tokenString inside the cookie's Value
	// for authentication before serving...
	// This is one big advantage of JWT over session. The burden has been shifted to client instead of taking memory space
	// on the server side. This helps a lot with the scaling process.

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Đăng nhập thành công",
		Data:       user,
	})
}

func (u *UserHandler) SignOut(c echo.Context) error {
	tcookie := http.Cookie{
		Name:     "AccessToken",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(&tcookie)

	rtcookie := http.Cookie{
		Name:     "RefreshToken",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(&rtcookie)

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Đăng xuất thành công",
	})

}

func (u *UserHandler) Profile(c echo.Context) error {
	tokenData := c.Get("user").(*jwt.Token)
	claims := tokenData.Claims.(*model.JwtCustomClaims)
	user, err := u.UserRepo.SelectUserById(c.Request().Context(), claims.UserId)

	if err != nil {
		if err == custom_error.UserNotFound {
			return c.JSON(http.StatusNotFound, model.Response{
				StatusCode: http.StatusNotFound,
				Message:    err.Error(),
				Data:       nil,
			})
		}

		return c.JSON(http.StatusUnauthorized, model.Response{
			StatusCode: http.StatusUnauthorized,
			Message:    err.Error(),
			Data:       nil,
		})
	}

	return c.JSON(http.StatusOK, model.Response{
		StatusCode: http.StatusOK,
		Message:    "Xử lý thành công",
		Data:       user,
	})
}

func (u *UserHandler) UpdateProfile(c echo.Context) error {
	req := requests.RequestUpdateUser{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	// validate requests
	err := c.Validate(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.Response{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	hash := security.HashAndSalt([]byte(req.Password))

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*model.JwtCustomClaims)
	user := model.User{
		UserId:   claims.UserId,
		FullName: req.FullName,
		Email:    req.Email,
		Password: hash,
	}

	user, err = u.UserRepo.UpdateUser(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model.Response{
		StatusCode: http.StatusCreated,
		Message:    "Xử lý thành công",
		Data:       user,
	})

}

func (u *UserHandler) RefeshToken(c echo.Context) error {
	cookie, err := c.Cookie("RefreshToken")
	if err != nil {
		return err
	}

	refreshCookie := cookie.Value

	token, err := jwt.Parse(refreshCookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Phương thức ký bất thường")
		}
		return []byte(security.SECRET_KEY), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		strClaims := fmt.Sprintf("%v", claims["UserId"])
		user, err := u.UserRepo.SelectUserById(c.Request().Context(), strClaims)
		if err != nil {
			log.Error(err)
			if err == custom_error.UserNotFound {
				return c.JSON(http.StatusNotFound, model.Response{
					StatusCode: http.StatusNotFound,
					Message:    err.Error(),
					Data:       nil,
				})
			}

			return c.JSON(http.StatusInternalServerError, model.Response{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
				Data:       nil,
			})
		}

		if strClaims == user.UserId {
			newAccessToken, err := security.CreateAccessToken(user)
			if err != nil {
				log.Error(err)
				return c.JSON(http.StatusInternalServerError, model.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Error(),
					Data:       nil,
				})
			}
			user.AccessToken = newAccessToken

			// create the access cookie for client(browser)
			newATCookie := &http.Cookie{
				Name:     "AccessToken",
				Value:    newAccessToken,
				Expires:  time.Now().Add(10 * time.Minute),
				HttpOnly: true,
			}
			c.SetCookie(newATCookie)

			newRefreshToken, err := security.CreateRefeshToken(user)
			if err != nil {
				log.Error(err)
				return c.JSON(http.StatusInternalServerError, model.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Error(),
					Data:       nil,
				})
			}
			user.RefreshToken = newRefreshToken

			// create the refresh cookie for client(browser)
			newRTCookie := &http.Cookie{
				Name:     "RefreshToken",
				Value:    newRefreshToken,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			c.SetCookie(newRTCookie)

			return c.JSON(http.StatusOK, model.Response{
				StatusCode: http.StatusOK,
				Message:    "Xử lý thành công",
				Data:       user,
			})

		}

		return c.JSON(http.StatusNotFound, model.Response{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
	}

	return c.JSON(http.StatusUnauthorized, model.Response{
		StatusCode: http.StatusUnauthorized,
		Message:    err.Error(),
	})

}
