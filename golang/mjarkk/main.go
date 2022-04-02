package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mjarkk/go-graphql"
	"io/ioutil"
	"log"
	"mime/multipart"
)

type QueryRoot struct{}

func (QueryRoot) ResolveHello() string {
	return "world"
}

type MethodRoot struct{}

func Handler() gin.HandlerFunc {
	graphqlSchema := graphql.NewSchema()
	err := graphqlSchema.Parse(QueryRoot{}, MethodRoot{}, nil)
	if err != nil {
		log.Fatal(err)
	}

	return func(c *gin.Context) {
		var form *multipart.Form
		getForm := func() (*multipart.Form, error) {
			if form != nil {
				return form, nil
			}
			return c.MultipartForm()
		}
		res, _ := graphqlSchema.HandleRequest(
			c.Request.Method,
			c.Query,
			func(key string) (string, error) {
				form, err := getForm()
				if err != nil {
					return "", err
				}
				values, ok := form.Value[key]
				if !ok || len(values) == 0 {
					return "", nil
				}
				return values[0], nil
			},
			func() []byte {
				requestBody, _ := ioutil.ReadAll(c.Request.Body)
				return requestBody
			},
			c.ContentType(),
			&graphql.RequestOptions{
				GetFormFile: func(key string) (*multipart.FileHeader, error) {
					form, err := getForm()
					if err != nil {
						return nil, err
					}
					files, ok := form.File[key]
					if !ok || len(files) == 0 {
						return nil, nil
					}
					return files[0], nil
				},
				Tracing: true,
			},
		)
		c.Data(200, "application/json", res)
	}
}

func main() {
	r := gin.New()
	r.POST("/graphql", Handler())

	fmt.Println("Now server is running on port 8080")
	fmt.Println("Test with POST      : curl -g --request POST 'http://localhost:8080/graphql?query={hello}'")
	r.Run() // listen and serve on 0.0.0.0:8080
}
