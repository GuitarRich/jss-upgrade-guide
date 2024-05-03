package main

import (
	"io"
	"strconv"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Contact struct {
	Id    int
	Name  string
	Email string
}

func newContact(id int, name, email string) Contact {
	return Contact{Id: id, Name: name, Email: email}
}

type Contacts []Contact

type Data struct {
	Contacts Contacts
}

func (d *Data) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.Id == id {
			return i
		}
	}

	return -1
}

func (d *Data) hasEmail(email string) bool {
	for _, c := range d.Contacts {
		if c.Email == email {
			return true
		}
	}

	return false
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact(1, "John", "john@example.com"),
			newContact(2, "Jane", "jane@example.com"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())

	// little bit of middlewards for housekeeping
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(20),
	)))

	e.Static("/images", "images")
	e.Static("/dist", "dist")

	page := newPage()
	id := 3
	e.Renderer = newTemplates()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			return c.Render(422, "form", formData)
		}

		contact := newContact(id, name, email)
		id++
		page.Data.Contacts = append(page.Data.Contacts, contact)
		c.Render(200, "form", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)

		if err != nil {
			return c.String(400, "Id must be an integer")
		}

		index := page.Data.indexOf(id)

		if index == -1 {
			return c.String(404, "Contact not found")
		}

		time.Sleep(4 * time.Second)
		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
