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

type Product struct {
	Asin   string `json:"asin"`
	Name   string `json:"name" validate:"required"`
	Maker  string `json:"maker" validate:"required"`
	Price  int    `json:"price" validate:"max=150"`
	Reason string `json:"reason" validate:"required"`
	Url    string `json:"url" validate:"required"`
}

type ProductPatch struct {
	Asin   *string `json:"asin"`
	Name   *string `json:"name"`
	Maker  *string `json:"maker"`
	Price  *int    `json:"price" validate:"max=150"`
	Reason *string `json:"reason"`
	Url    *string `json:"url"`
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
	dsn := "docker:docker@tcp(127.0.0.1:3307)/AmazonApi?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	// Migrate the schema,product.go参照
	if err := db.AutoMigrate(&repository.Products{}); err != nil {
		panic(err.Error())
	}

	e.POST("/amazon", func(context echo.Context) error {
		// リクエストを取得する
		product := new(Product)
		_ = context.Bind(product)
		// バリデーション
		if err := context.Validate(product); err != nil {
			return context.String(http.StatusBadRequest, err.Error())
		}
		// Create
		now := time.Now()
		db.Create(&repository.Products{
			Asin:      product.Asin,
			Name:      product.Name,
			Maker:     product.Maker,
			Price:     product.Price,
			Reason:    product.Reason,
			Url:       product.Url,
			Status:    false,
			CreatedAt: now,
			UpdatedAt: now,
		})
		return context.JSON(http.StatusCreated, product)
	})

	e.PATCH("/amazon/inactive/:asin", func(context echo.Context) error {
		// リクエストを取得する
		asin := context.Param("asin")

		m := new(repository.Products)
		// First
		if tx := db.First(m, "asin = ?", asin); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}

		m.Status = true
		if tx := db.Model(m).Where("asin = ?", asin).Updates(m); tx.Error != nil {
			return context.String(http.StatusBadRequest, tx.Error.Error())
		}

		return context.JSON(http.StatusAccepted, nil)
	})

	e.GET("/amazon/:asin", func(context echo.Context) error {
		// リクエストを取得する
		asin := context.Param("asin")
		m := new(repository.Products)
		if tx := db.First(m, "asin = ?", asin); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}

		product := &Product{
			Asin:   m.Asin,
			Name:   m.Name,
			Maker:  m.Maker,
			Price:  m.Price,
			Reason: m.Reason,
			Url:    m.Url,
		}
		return context.JSON(http.StatusOK, product)
	})
	e.PUT("/amazon/:asin", func(context echo.Context) error {
		// リクエストを取得する
		asin := context.Param("asin")
		product := new(Product)
		_ = context.Bind(product)

		// バリデーション
		if err := context.Validate(product); err != nil {
			return context.JSON(http.StatusBadRequest, err)
		}

		m := new(repository.Products)
		// First
		if tx := db.First(m, "asin = ?", asin); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}
		// Update
		now := time.Now()
		db.Model(m).
			Where("asin = ?", asin).
			Updates(repository.Products{
				Name:      product.Name,
				Maker:     product.Maker,
				Price:     product.Price,
				Reason:    product.Reason,
				UpdatedAt: now,
			})
		return context.JSON(http.StatusOK, product)
	})
	e.PATCH("/amazon/:asin", func(context echo.Context) error {
		// リクエストを取得する
		asin := context.Param("asin")
		product := new(ProductPatch)
		_ = context.Bind(product)
		// バリデーション
		if err := context.Validate(product); err != nil {
			return context.JSON(http.StatusBadRequest, err)
		}

		m := new(repository.Products)
		// First
		if tx := db.First(m, "asin = ?", asin); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}

		tx := db.Model(m).Where("asin = ?", asin)

		if product.Name != nil {
			m.Name = *product.Name
		}
		if product.Maker != nil {
			m.Name = *product.Maker
		}
		if product.Price != nil {
			m.Price = *product.Price
		}
		if product.Reason != nil {
			m.Reason = *product.Reason
		}
		if product.Url != nil {
			m.Url = *product.Url
		}
		tx.Updates(*m)
		return context.JSON(http.StatusOK, &Product{
			Asin:   m.Asin,
			Name:   m.Name,
			Maker:  m.Maker,
			Price:  m.Price,
			Reason: m.Reason,
			Url:    m.Url,
		})
	})
	e.DELETE("/amazon/:asin", func(context echo.Context) error {
		// リクエストを取得する
		asin := context.Param("asin")

		m := new(repository.Products)
		// First
		if tx := db.First(m, "asin = ?", asin); tx.Error != nil {
			return context.String(http.StatusNotFound, tx.Error.Error())
		}
		db.Delete(m, "asin = ?", asin)

		return context.String(http.StatusNoContent, "")
	})
	e.Logger.Fatal(e.Start(":1232"))
}
