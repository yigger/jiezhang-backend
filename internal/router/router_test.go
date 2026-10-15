package router

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRouteContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	Register(engine, func(c *gin.Context) { c.Next() }, Handlers{})
	var routes []string
	for _, r := range engine.Routes() {
		if strings.HasPrefix(r.Path, "/api/") {
			routes = append(routes, r.Method+" "+r.Path)
		}
	}
	sort.Strings(routes)
	expected, err := os.ReadFile("testdata/routes.golden")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(routes, "\n")+"\n" != string(expected) {
		t.Fatalf("route contract changed:\n%s", strings.Join(routes, "\n"))
	}
}
