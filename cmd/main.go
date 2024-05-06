package main

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/hashicorp/go-version"
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

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Version struct {
	Id      string
	Version string
}

type UpgradeStep struct {
	Id      string
	Version string
	Steps   string
}

type PageData struct {
	Title        string
	Versions     []*version.Version
	UpgradeSteps []UpgradeStep
}

func newPageData(versions []*version.Version) PageData {
	return PageData{
		Title:    "GuitarRich: JSS SDK Upgrade Guide",
		Versions: versions,
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

	e.Renderer = newTemplate()

	e.Static("/images", "images")
	e.Static("/dist", "dist")

	e.GET("/", func(c echo.Context) error {
		pageData := newPageData(getVersions())
		fmt.Printf("pageData: %+v\n", pageData)
		return c.Render(200, "index", pageData)
	})

	e.POST("/api/upgrade-steps", func(c echo.Context) error {
		fmt.Printf("POST/api/upgrade-steps\n")
		startingVersionParam := c.FormValue("starting-version")
		targetVersionParam := c.FormValue("target-version")
		fmt.Printf("startignVersionParam: %s\n", startingVersionParam)
		fmt.Printf("targetVersionParam: %s\n", targetVersionParam)

		if (startingVersionParam == "") || (targetVersionParam == "") {
			return c.String(422, "missing starting or target version")
		}

		startingVersion, _ := version.NewVersion(startingVersionParam)
		targetVersion, _ := version.NewVersion(targetVersionParam)

		fmt.Printf("startingVersion: %s\n", startingVersion)
		fmt.Printf("targetVersion: %s\n", targetVersion)

		if startingVersion.GreaterThan(targetVersion) {
			return c.String(422, "starting version must be less than target version")
		}

		upgradeSteps := getInstructions()

		return c.Render(200, "steps", upgradeSteps)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "42069"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
