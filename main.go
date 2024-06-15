package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	trmgorm "github.com/avito-tech/go-transaction-manager/gorm"
	"github.com/gofiber/fiber"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Course struct {
	Id   uuid.UUID `gorm:"primaryKey;type:uuid;column:course_id"`
	Name string    `gorm:"unique;type:varchar;column:name"`
	URL  string    `gorm:"type:varchar;column:url"`
	Text string    `gorm:"type:text;column:text"`
}

// DTO COURSE

type CreateCourseInput struct {
	Name string `validate:"required" json:"name"`
	URL  string `validate:"required" json:"url"`
	Text string `validate:"required" json:"text"`
}

type GetCourseInput struct {
	Id string `validate:"required" json:"id"`
}

type DeleteCourseInput struct {
	Id string `validate:"required" json:"id"`
}

type UpdateCourseInput struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Text string `json:"text"`
}

type GetAllCoursesInput struct {
	Ids []string `validate:"required" json:"ids"`
}

// ------------------------------------------------------------------------

type User struct {
	Id       uuid.UUID `gorm:"primaryKey;type:uuid;column:user_id"`
	Email    string    `gorm:"unique;type:varchar;column:email"`
	Password string    `gorm:"type:varchar;column:password"`
	Role     bool      `gorm:"type:boolean;column:role"`
}

// DTO USER

type AuthInput struct {
	Email    string `validate:"required" json:"email"`
	Password string `validate:"required" json:"password"`
	Role     bool   `validate:"required" json:"role"`
}

type SignInInput struct {
	Email    string `validate:"required" json:"email"`
	Password string `validate:"required" json:"password"`
	Role     bool   `validate:"required" json:"role"`
}

type Error struct {
	Message string `json:"message"`
}

var (
	ErrNotFound          = errors.New("Entity not found")
	ErrAlreadyExist      = errors.New("Entity already exists")
	ErrUnknown           = errors.New("Unknown error")
	ErrInvalidEntity     = errors.New("Invalid entity")
	ErrInvalidField      = errors.New("Invalid field")
	ErrInvalidSQLRequest = errors.New("Invalid SQL request")
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DBNAME")
	dbSSL := os.Getenv("DB_SSL")

	// Connect to the database
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL)), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Course{}, &User{})

	app := fiber.New()

	// USER

	app.Get("/", func(c *fiber.Ctx) {
		c.Send("Hello, World!")
	})

	app.Post("/registraition", func(c *fiber.Ctx) {
		var req AuthInput
		if err := c.BodyParser(&req); err != nil {
			BadRequest(c, "Invalid request")
		}

		res, err := CreateUser(context.Background(), &User{
			Id:       uuid.New(),
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
		}, db, trmgorm.DefaultCtxGetter)

		if err != nil {
			switch {
			case errors.Is(err, ErrAlreadyExist):
				BadRequest(c, "Already exists")
			default:
				Internal(c, "Unknown error")
			}
		}

		OK(c, res)
	})

	app.Post("/signIn", func(c *fiber.Ctx) {
		var req SignInInput
		if err := c.BodyParser(&req); err != nil {
			BadRequest(c, "Invalid request")
		}

		res, err := Auth(context.Background(), req.Email, db, trmgorm.DefaultCtxGetter)

		if err != nil {
			switch {
			case errors.Is(err, ErrNotFound):
				BadRequest(c, "NotFound")
			default:
				Internal(c, "Unknown error")
			}
		}

		if req.Email == res.Email && req.Password == res.Password {
			OK(c, res)
		} else {
			BadRequest(c, "Invalid email or password")
		}
	})

	// COURSE

	app.Post("/createCourse", func(c *fiber.Ctx) {
		var req CreateCourseInput
		if err := c.BodyParser(&req); err != nil {
			BadRequest(c, "Invalid request")
		}

		res, err := CreateCourse(context.Background(), &Course{
			Id:   uuid.New(),
			Name: req.Name,
			URL:  req.URL,
			Text: req.Text,
		}, db, trmgorm.DefaultCtxGetter)

		if err != nil {
			switch {
			case errors.Is(err, ErrAlreadyExist):
				BadRequest(c, "Already exists")
			default:
				Internal(c, "Unknown error")
			}
		}

		OK(c, res)
	})

	app.Post("/updateCourse", func(c *fiber.Ctx) {
		var req UpdateCourseInput
		if err := c.BodyParser(&req); err != nil {
			BadRequest(c, "Invalid request")
		}

		id, err := uuid.Parse(req.Id)
		if err != nil {
			BadRequest(c, "Invalid id")
		}

		res, err := UpdateCourse(context.Background(), &Course{
			Id:   id,
			Name: req.Name,
			URL:  req.URL,
			Text: req.Text,
		}, db, trmgorm.DefaultCtxGetter)

		if err != nil {
			switch {
			case errors.Is(err, ErrAlreadyExist):
				BadRequest(c, "Already exists")
			default:
				Internal(c, "Unknown error")
			}
		}

		OK(c, res)
	})

	app.Post("/deleteCourse", func(c *fiber.Ctx) {
		var req DeleteCourseInput
		if err := c.BodyParser(&req); err != nil {
			BadRequest(c, "Invalid request")
		}

		id, err := uuid.Parse(req.Id)
		if err != nil {
			BadRequest(c, "Invalid id")
		}

		res, err := DeleteCourse(context.Background(), id, db, trmgorm.DefaultCtxGetter)

		if err != nil {
			switch {
			case errors.Is(err, ErrNotFound):
				BadRequest(c, "ErrNotFound")
			default:
				Internal(c, "Unknown error")
			}
		}

		OK(c, res)
	})

	app.Post("/getCourse", func(c *fiber.Ctx) {
		var req GetCourseInput
		if err := c.BodyParser(&req); err != nil {
			BadRequest(c, "Invalid request")
		}

		id, err := uuid.Parse(req.Id)
		if err != nil {
			BadRequest(c, "Invalid id")
		}

		res, err := DeleteCourse(context.Background(), id, db, trmgorm.DefaultCtxGetter)

		if err != nil {
			switch {
			case errors.Is(err, ErrNotFound):
				BadRequest(c, "ErrNotFound")
			default:
				Internal(c, "Unknown error")
			}
		}

		OK(c, res)
	})

	app.Post("/getCourses", func(c *fiber.Ctx) {
		res, err := GetCourses(context.Background(), db, trmgorm.DefaultCtxGetter)

		if err != nil {
			Internal(c, "Unknown error")
		}

		OK(c, res)
	})

	app.Listen("0.0.0.0:8080")
}

// BD COURSE

func CreateCourse(ctx context.Context, course *Course, db *gorm.DB, getter *trmgorm.CtxGetter) (*Course, error) {
	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	err := tr.Create(course).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return nil, ErrAlreadyExist
		default:
			return nil, err
		}
	}

	return course, nil
}

func GetCourse(ctx context.Context, courseId uuid.UUID, db *gorm.DB, getter *trmgorm.CtxGetter) (*Course, error) {
	course := new(Course)

	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	err := tr.First(course, "course_id = ?", courseId).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return course, nil
}

func GetCourses(ctx context.Context, db *gorm.DB, getter *trmgorm.CtxGetter) ([]*Course, error) {
	user := make([]*Course, 0, 0)

	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	err := tr.Find(&user).Error

	if err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateCourse(ctx context.Context, course *Course, db *gorm.DB, getter *trmgorm.CtxGetter) (*Course, error) {
	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)

	result := tr.Model(course).Where("course_id = ?", course.Id).Updates(map[string]interface{}{"name": course.Name, "url": course.URL, "text": course.Text})

	if result.Error != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, result.Error
		}
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return course, nil
}

func DeleteCourse(ctx context.Context, userId uuid.UUID, db *gorm.DB, getter *trmgorm.CtxGetter) (*Course, error) {
	course := new(Course)

	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	result := tr.Clauses(clause.Returning{}).Where("course_id = ?", userId).Delete(course)
	if result.Error != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, result.Error
		}
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return course, nil
}

// BD USER

func CreateUser(ctx context.Context, user *User, db *gorm.DB, getter *trmgorm.CtxGetter) (*User, error) {
	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	err := tr.Create(user).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return nil, ErrAlreadyExist
		default:
			return nil, err
		}
	}

	return user, nil
}

func GetUser(ctx context.Context, userId uuid.UUID, db *gorm.DB, getter *trmgorm.CtxGetter) (*User, error) {
	user := new(User)

	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	err := tr.First(user, "user_id = ?", userId).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func Auth(ctx context.Context, email string, db *gorm.DB, getter *trmgorm.CtxGetter) (*User, error) {
	user := new(User)

	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	err := tr.First(user, "email = ?", email).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func UpdateUser(ctx context.Context, user *User, db *gorm.DB, getter *trmgorm.CtxGetter) (*User, error) {
	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)

	result := tr.Model(user).Where("user_id = ?", user.Id).Updates(map[string]interface{}{"email": user.Email, "password": user.Password})

	if result.Error != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, result.Error
		}
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return user, nil
}

func DeleteUser(ctx context.Context, userId uuid.UUID, db *gorm.DB, getter *trmgorm.CtxGetter) (*User, error) {
	user := new(User)

	tr := getter.DefaultTrOrDB(ctx, db).WithContext(ctx)
	result := tr.Clauses(clause.Returning{}).Where("user_id = ?", userId).Delete(user)
	if result.Error != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrRecordNotFound):
			return nil, ErrNotFound
		default:
			return nil, result.Error
		}
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return user, nil
}

// HTTP RESPONSE

func BadRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(&Error{Message: message})
}

func Internal(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(&Error{Message: message})
}
func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(&Error{Message: message})
}

func OK(c *fiber.Ctx, data any) error {
	if data == nil {
		return c.Status(fiber.StatusOK).JSON(&Error{Message: "OK"})
	}

	return c.Status(fiber.StatusOK).JSON(data)
}
