package controller

import (
	appErrors "github.com/HieuNguyenVV/sp-agent/internal/errors"
	"github.com/HieuNguyenVV/sp-agent/internal/server/validator"
	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	_ = c.Error(err)

	appError, ok := err.(appErrors.AppError)
	if ok {
		c.AbortWithStatusJSON(appError.StatusCode, err)
		return
	}
	c.AbortWithStatusJSON(appErrors.ErrInternalServerError.StatusCode, err)
}

func HandleBindError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	HandleError(c, validator.BindError(err))
}
