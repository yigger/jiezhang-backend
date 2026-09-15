package router

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSwaggerMatchesRouteContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterDocumentation(engine)
	for _, path := range []string{"/swagger/index.html", "/swagger/doc.json"} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 200 {
			t.Fatalf("%s: status %d", path, rec.Code)
		}
		if path != "/swagger/doc.json" {
			if !strings.Contains(rec.Body.String(), "Swagger UI") {
				t.Fatal("missing UI")
			}
			continue
		}
		var spec struct {
			BasePath string `json:"basePath"`
			Paths    map[string]map[string]struct {
				ID         string                     `json:"operationId"`
				Responses  map[string]json.RawMessage `json:"responses"`
				Parameters []struct {
					In       string `json:"in"`
					Name     string `json:"name"`
					Required bool   `json:"required"`
				} `json:"parameters"`
			} `json:"paths"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
			t.Fatal(err)
		}
		var routes []string
		ids := map[string]bool{}
		params := regexp.MustCompile(`\{([^}]+)\}`)
		for path, methods := range spec.Paths {
			for method, op := range methods {
				if op.ID == "" || ids[op.ID] {
					t.Errorf("missing/duplicate operation ID: %s %s", method, path)
				}
				ids[op.ID] = true
				if len(op.Responses) == 0 {
					t.Errorf("missing responses: %s", op.ID)
				}
				for _, match := range params.FindAllStringSubmatch(path, -1) {
					found := false
					for _, p := range op.Parameters {
						if p.In == "path" && p.Name == match[1] && p.Required {
							found = true
						}
					}
					if !found {
						t.Errorf("missing required path parameter %s: %s", match[1], op.ID)
					}
				}
				routes = append(routes, strings.ToUpper(method)+" "+spec.BasePath+params.ReplaceAllString(path, ":$1"))
			}
		}
		sort.Strings(routes)
		want, err := os.ReadFile("testdata/routes.golden")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(routes, "\n")+"\n" != string(want) {
			t.Fatal("Swagger operations differ from the API route baseline")
		}
	}
}

func TestSwaggerDirectoryRedirectsToUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterDocumentation(engine)
	for _, path := range []string{"/swagger", "/swagger/"} {
		t.Run(path, func(t *testing.T) {
			current := path
			for redirects := 0; redirects < 3; redirects++ {
				rec := httptest.NewRecorder()
				engine.ServeHTTP(rec, httptest.NewRequest("GET", current, nil))
				if rec.Code == 200 {
					if current != "/swagger/index.html" || !strings.Contains(rec.Body.String(), "Swagger UI") {
						t.Fatalf("unexpected destination %s", current)
					}
					return
				}
				if rec.Code != 301 && rec.Code != 302 {
					t.Fatalf("%s: status %d", current, rec.Code)
				}
				current = rec.Header().Get("Location")
			}
			t.Fatal("too many redirects")
		})
	}
}
