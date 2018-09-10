package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/julienschmidt/httprouter"
)

// Web Client
// Handle assets and client app preprarations

// Prepare client HTML:
// Set paths and add version to assets.
func webPrepareClientHtml(html string) string {
	status, _ := NewAppStatus()

	// Replace paths and tags
	rewriter := strings.NewReplacer(
		// Paths
		"js/", "/static/js/",
		"css/", "/static/css/",

		// Tags
		"APP_VERSION", status.Version,
	)
	html = rewriter.Replace(html)
	return html
}

// Register assets handler and index handler
// at /static and /
func webRegisterAssets(ui UiConfig, router *httprouter.Router) error {
	log.Println("Preparing and installing assets")

	// Serve static assets
	assets := rice.MustFindBox("../client/build")
	assetsHandler := http.StripPrefix(
		"/static/",
		http.FileServer(assets.HTTPBox()))

	// Prepare client html: Rewrite paths
	indexHtml, err := assets.String("index.html")
	if err != nil {
		return err
	}

	theme := NewTheme(ui.Theme)
	err = theme.RegisterThemeAssets(router)
	if err != nil {
		log.Println("Warning:", err)
	}

	// Update paths
	indexHtml = webPrepareClientHtml(indexHtml)

	// Register static assets
	router.Handler("GET", "/static/*path", assetsHandler)

	// Rewrite paths
	// Serve index html as root...
	router.GET("/",
		func(res http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			// Include theme, we need to update the
			// hashes on reload, so we can check if the theme has
			// changed without restarting the app
			themedHtml := theme.PrepareClientHtml(indexHtml)
			io.WriteString(res, themedHtml)
		})

	// ...and as catch all
	router.GET("/alice/*path",
		func(res http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
			// ditto here
			themedHtml := theme.PrepareClientHtml(indexHtml)
			io.WriteString(res, themedHtml)
		})

	return nil
}
