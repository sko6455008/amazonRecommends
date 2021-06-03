package main

import (
	"net/http"
	"simpleApi/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	Code string `json:"code"`
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"max=150"`
}

type UserPatch struct {
	Code *string `json:"code"`
	Name *string `json:"name"`
	Age  *int    `json:"age" validate:"max=150"`
}

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Validator = NewValidator()

	//mysql connection
	dsn := "docker:docker@tcp(127.0.0.1:3307)/sampleApi?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	// Migrate the schema,user.go参照
	if err := db.AutoMigrate(&repository.Users{}); err != nil {
		panic(err.Error())
	}

	e.POST("/simple", func(context echo.Context) error {
		// リクエストを取得する
		user := new(User)
		_ = context.Bind(user)
		// バリデーション
		if err := context.Validate(user); err != nil {
			return context.String(http.StatusBadRequest, err.Error())
		}
		// Create
		now := time.Now()
		db.Create(&repository.Users{
			Code:      user.Code,
			Name:      user.Name,
			Age:       user.Age,
			CreatedAt: now,
			UpdatedAt: now,
		})
		return context.JSON(http.StatusCreated, user)
	})

	e.GET("/simple/:code", func(context echo.Context) error {
		// リクエストを取得する
		code := context.Param("code")
		m := new(repository.Users)
		if tx := db.First(m, "code = ?", code); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}

		user := &User{
			Code: m.Code,
			Name: m.Name,
			Age:  m.Age,
		}
		return context.JSON(http.StatusOK, user)
	})
	e.PUT("/simple/:code", func(context echo.Context) error {
		// リクエストを取得する
		code := context.Param("code")
		user := new(User)
		_ = context.Bind(user)

		// バリデーション
		if err := context.Validate(user); err != nil {
			return context.JSON(http.StatusBadRequest, err)
		}

		m := new(repository.Users)
		// First
		if tx := db.First(m, "code = ?", code); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}
		// Update
		now := time.Now()
		db.Model(m).
			Where("code = ?", code).
			Updates(repository.Users{
				Name:      user.Name,
				Age:       user.Age,
				UpdatedAt: now,
			})
		return context.JSON(http.StatusOK, user)
	})
	e.PATCH("/simple/:code", func(context echo.Context) error {
		// リクエストを取得する
		code := context.Param("code")
		user := new(UserPatch)
		_ = context.Bind(user)
		// バリデーション
		if err := context.Validate(user); err != nil {
			return context.JSON(http.StatusBadRequest, err)
		}

		m := new(repository.Users)
		// First
		if tx := db.First(m, "code = ?", code); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}

		tx := db.Model(m).Where("code = ?", code)
		if user.Age != nil {
			m.Age = *user.Age
		}
		if user.Name != nil {
			m.Name = *user.Name
		}
		tx.Updates(*m)
		return context.JSON(http.StatusOK, &User{
			Code: m.Code,
			Name: m.Name,
			Age:  m.Age,
		})
	})
	e.DELETE("/simple/:code", func(context echo.Context) error {
		// リクエストを取得する
		code := context.Param("code")

		m := new(repository.Users)
		// First
		if tx := db.First(m, "code = ?", code); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}
		db.Delete(m, "code = ?", code)

		return context.String(http.StatusNoContent, "")
	})
	e.Logger.Fatal(e.Start(":1232"))
}
