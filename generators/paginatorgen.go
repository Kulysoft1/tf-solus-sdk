// +build ignore

/*
	Generator for common paginated response's methods.

	For now will add only `Next(ctx context.Context) bool` method to the collection
	response which can be used for iteration through all available entities.

	Usage:
	Add this line to one of package file:

		//go:generate go run generator/paginatorgen.go

	Add the `paginatedResponse` type and `Data` field to the required structure.
	This field should be a slice of entity type.

	Example:

		type UsersResponse struct {
			paginatedResponse

			Data []User `json:"data"`
		}

	By this generator we add generic types to Golang.
*/
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Cannot generate paginated data methods: %s", err)
		os.Exit(1)
	}
}

func run() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()
		if info.IsDir() {
			if shouldIgnoreDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnoreFile(name) {
			return nil
		}

		return processFile(path)
	})
}

func shouldIgnoreDir(name string) bool {
	return (name == "vendor") || (name[0] == '.')
}

func shouldIgnoreFile(name string) bool {
	return !strings.HasSuffix(name, ".go") ||
		strings.HasSuffix(name, "_test.go") ||
		strings.HasSuffix(name, "_gen.go")
}

func processFile(p string) error {
	fmt.Printf("Process file %q\n", p)

	data, err := collectStructData(p)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	dir := filepath.Dir(p)
	name := filepath.Base(p)
	// We check that file should have ".go" suffix, so `LastIndex` will never returns
	// -1 here.
	name = name[:strings.LastIndex(name, ".")]

	err = renderCode(filepath.Join(dir, fmt.Sprintf("%s_gen.go", name)), data)
	if err != nil {
		return err
	}
	return renderTests(filepath.Join(dir, fmt.Sprintf("%s_gen_test.go", name)), data)
}

func collectStructData(p string) ([]structData, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, p, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	strctData := make([]structData, 0)
	for _, d := range f.Decls {
		decl, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		name := ""
		dataType := ""

		for _, s := range decl.Specs {
			spec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typ, ok := spec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Make sure that current structure is suitable for us.
			for _, field := range typ.Fields.List {
				ident, ok := field.Type.(*ast.Ident)
				if !ok {
					continue
				}

				if field.Names != nil || ident.Name != "paginatedResponse" {
					continue
				}

				name = spec.Name.Name
				break
			}

			if name == "" {
				break
			}

			// Make sure that current structure has correct Data field
			for _, field := range typ.Fields.List {
				if field.Names == nil {
					continue
				}

				found := false
				for _, n := range field.Names {
					if n.Name == "Data" {
						found = true
						break
					}
				}

				if !found {
					continue
				}

				arrTyp, ok := field.Type.(*ast.ArrayType)
				if !ok {
					return nil, fmt.Errorf("paginated response %q has Data field but it not an array", name)
				}

				ident, ok := arrTyp.Elt.(*ast.Ident)
				if !ok {
					continue
				}

				dataType = ident.Name
			}
		}

		if name == "" {
			continue
		}

		if dataType == "" {
			return nil, fmt.Errorf("structure %q has paginatedRespnse but doesn't have Data field", name)
		}

		strctData = append(strctData, structData{
			// For now we assume that all response structure will have `r`
			// receiver.
			// Of course we could parse whole file and try to find at least
			// one method and get the receiver ... but I'm a bit lazy to
			// do it :)
			Receiver:   "r",
			Name:       name,
			Entrypoint: strings.ToLower(strings.TrimSuffix(name, "Response")),
			DataType:   dataType,
		})
	}
	return strctData, nil
}

type structData struct {
	Receiver   string
	Name       string
	Entrypoint string
	DataType   string
}

func renderCode(p string, data []structData) error {
	//language=GoTemplate
	return renderFileTemplate(p, `// Autogenerated file. Do not edit!

package solus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)
{{ range . }}
// Next using for iterating through all data entities.
//
// Examples:
//
//	ctx, cancelFunc := context.WithTimeout(context.Background(), 30 * time.Second)
//	defer cancelFunc()
//
//  for {
//		for _, datum := range resp.Data {
//			doSmthWithDatum(datum)
//		}
//
//		if !resp.Next(ctx) {
//			break
//		}
//	}
//
//  if resp.Err() != nil {
//		handleAnError(resp.Err())
//	}
func ({{ .Receiver }} *{{ .Name }}) Next(ctx context.Context) bool {
	if ({{ .Receiver }}.Meta.LastPage == {{ .Receiver }}.Meta.CurrentPage) || ({{ .Receiver }}.err != nil) {
		return false
	}

	body, code, err := {{ .Receiver }}.service.client.request(ctx, http.MethodGet, {{ .Receiver }}.Links.Next)
	if err != nil {
		{{ .Receiver }}.err = err
		return false
	}

	if code != http.StatusOK {
		{{ .Receiver }}.err = newHTTPError(http.MethodGet, {{ .Receiver }}.Links.Next, code, body)
		return false
	}

	if err := json.Unmarshal(body, &{{ .Receiver }}); err != nil {
		{{ .Receiver }}.err = fmt.Errorf("failed to decode %q: %s", body, err)
		return false
	}
	return true
}
{{ end }}
`, data)
}

func renderTests(p string, data []structData) error {
	//language=GoTemplate
	return renderFileTemplate(p, `// Autogenerated file. Do not edit!

package solus

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
{{ range . }}
func Test{{ .Name }}_Next(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		page := int32(1)

		s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			p := atomic.LoadInt32(&page)

			assert.Equal(t, "/{{ .Entrypoint }}", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, strconv.Itoa(int(p)), r.URL.Query().Get("page"))

			if p == 3 {
				writeJSON(t, w, http.StatusOK, {{ .Name }}{
					Data: []{{ .DataType }}{
						{
							ID: int(p),
						},
					},
					paginatedResponse: paginatedResponse{
						Links: ResponseLinks{
							Next: r.URL.String(),
						},
						Meta: ResponseMeta{
							CurrentPage: int(p),
							LastPage: 3,
						},
					},
				})
				return
			}
			atomic.AddInt32(&page, 1)

			q := r.URL.Query()
			q.Set("page", strconv.Itoa(int(p) + 1))
			r.URL.RawQuery = q.Encode()

			writeJSON(t, w, http.StatusOK, {{ .Name }}{
				paginatedResponse: paginatedResponse{
					Links: ResponseLinks{
						Next: r.URL.String(),
					},
					Meta: ResponseMeta{
						CurrentPage: int(p),
						LastPage: 3,
					},
				},
				Data: []{{ .DataType }}{{"{{"}}ID: int(p){{"}}"}},
			})
		})
		defer s.Close()

		resp := {{ .Name }}{
			paginatedResponse: paginatedResponse{
				Links: ResponseLinks{
					Next: fmt.Sprintf("%s/{{ .Entrypoint }}?page=1", s.URL),
				},
				Meta: ResponseMeta{
					CurrentPage: 1,
					LastPage: 3,
				},
				service: &service{createTestClient(t, s.URL)},
			},
		}

		i := 1
		for resp.Next(context.Background()) {
			require.Equal(t, []{{ .DataType }}{{"{{"}}ID: i{{"}}"}}, resp.Data)
			i++
		}
		require.NoError(t, resp.err)
		require.Equal(t, 4, i, "Expects to get 3 entity, but got less")
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("unexpected status code", func(t *testing.T) {
			s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/{{ .Entrypoint }}", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, strconv.Itoa(1), r.URL.Query().Get("page"))
				w.WriteHeader(http.StatusBadRequest)
			})
			defer s.Close()

			resp := {{ .Name }}{
				paginatedResponse: paginatedResponse{
					Links: ResponseLinks{
						Next: fmt.Sprintf("%s/{{ .Entrypoint }}?page=1", s.URL),
					},
					Meta: ResponseMeta{
						CurrentPage: 1,
						LastPage:    3,
					},
					service: &service{createTestClient(t, s.URL)},
				},
			}

			resp.Next(context.Background())
			assert.EqualError(t, resp.Err(), fmt.Sprintf(
				"HTTP GET %s/{{ .Entrypoint }}?page=1 returns 400 status code",
				s.URL,
			))
		})

		t.Run("failed to unmarshal", func(t *testing.T) {
			s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/{{ .Entrypoint }}", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, strconv.Itoa(1), r.URL.Query().Get("page"))

				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("fake"))
				require.NoError(t, err)
			})
			defer s.Close()

			resp := {{ .Name }}{
				paginatedResponse: paginatedResponse{
					Links: ResponseLinks{
						Next: fmt.Sprintf("%s/{{ .Entrypoint }}?page=1", s.URL),
					},
					Meta: ResponseMeta{
						CurrentPage: 1,
						LastPage:    3,
					},
					service: &service{createTestClient(t, s.URL)},
				},
			}

			resp.Next(context.Background())
			assert.EqualError(
				t,
				resp.Err(),
				"failed to decode \"fake\": invalid character 'k' in literal false (expecting 'l')",
			)
		})
	})
}
{{ end }}`, data)
}

func renderFileTemplate(p, tmpl string, data []structData) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("cannot gofmt code: %w", err)
	}

	return ioutil.WriteFile(p, src, 0644)
}
