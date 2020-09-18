package main

import (
	"crypto/subtle"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincent-petithory/dataurl"
)

var (
	flags struct {
		Port    int
		Root    string
		MaxSize int
		Key     string
	}
)

func renderErrorPage(c *gin.Context, s int, m string) {
	c.HTML(s, "error.gohtml", struct {
		Message string
	}{
		Message: m,
	})
}

func getNew(c *gin.Context) {
	if flags.Key == "" {
		renderErrorPage(c, http.StatusBadRequest, "Administrative mode is disabled.")
		return
	}

	qk := c.Query("key")
	if subtle.ConstantTimeCompare([]byte(qk), []byte(flags.Key)) == 0 {
		renderErrorPage(c, http.StatusUnauthorized, "Invalid administrative key.")
		return
	}

	i := uuid.New().String()
	f, err := os.Create(path.Join(flags.Root, i))
	if err != nil {
		renderErrorPage(c, http.StatusInternalServerError, "Creation failed.")
	}
	f.Close()
	c.Redirect(http.StatusTemporaryRedirect, "id/"+i)

	return
}

func getIDID(c *gin.Context) {
	id := c.Param("id")
	if strings.Contains("..", id) {
		c.Status(http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(path.Join(flags.Root, id)); err != nil {
		renderErrorPage(c, http.StatusInternalServerError, "Not found.")
		return
	}

	c.HTML(http.StatusOK, "index.gohtml", struct {
		ID           string
		MaxSize      int
		MaxSizeHuman string
	}{
		ID:           id,
		MaxSize:      flags.MaxSize,
		MaxSizeHuman: fmt.Sprintf("%.2f MiB", float64(flags.MaxSize)/1024/1024),
	})
}

func putIDID(c *gin.Context) {
	m := struct {
		ID      string
		Name    string
		Size    int
		DataURL string
	}{}

	mx := int64(flags.MaxSize * 2)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, mx)
	if err := c.BindJSON(&m); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if strings.Contains("..", m.ID) {
		c.Status(http.StatusBadRequest)
		return
	}

	d, err := dataurl.DecodeString(m.DataURL)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	n := path.Join(flags.Root, m.ID)
	if _, err := os.Stat(n); err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if len(d.Data) > flags.MaxSize {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := ioutil.WriteFile(n, d.Data, 0644); err != nil {
		c.Status(http.StatusInternalServerError)
	}
}

func noRoute(c *gin.Context) {
	c.HTML(http.StatusNotFound, "error.gohtml", struct {
		Message string
	}{
		Message: "Not found.",
	})
}

func main() {
	flag.IntVar(&flags.Port, "p", 1501, "Port")
	flag.StringVar(&flags.Root, "r", "./files", "Files root path")
	flag.IntVar(&flags.MaxSize, "m", 100*1024*1024, "Maximum size of uploaded files")
	flag.StringVar(&flags.Key, "k", "key", "Administrative key. Administration is disabled if left empty")
	flag.Parse()

	if fi, err := os.Stat(flags.Root); err != nil || !fi.IsDir() {
		log.Fatalf("failed to stat files root: %s", err)
	}

	g := gin.Default()
	g.Delims("{{{", "}}}")
	g.LoadHTMLGlob("templates/*")
	g.GET("/new", getNew)
	g.GET("/id/:id", getIDID)
	g.PUT("/id/:id", putIDID)
	g.Static("static", "static/")
	g.NoRoute(noRoute)
	g.Run(fmt.Sprintf(":%d", flags.Port))
}
