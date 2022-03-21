package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/marceloSantosC/go-server"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const GoogleBooksBaseUrl = "https://www.googleapis.com/books"

type BookFilters struct {
	Language   string `json:"language"`
	Tittle     string `json:"tittle"`
	Author     string `json:"author"`
	Subject    string `json:"subject"`
	StartIndex int    `json:"startIndex"`
}

type Book struct {
	Language      string   `json:"language"`
	Tittle        string   `json:"tittle"`
	Authors       []string `json:"authors"`
	Subjects      []string `json:"subjects"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"publishedDate"`
	NumberOfPages int      `json:"numberOfPages"`
	Description   string   `json:"description"`
	Ebook         bool     `json:"ebook"`
	PublicDomain  bool     `json:"publicDomain"`
	LinkToBuy     string   `json:"LinkToBuy"`
}

type GBooksResponse struct {
	TotalItems int64                 `json:"totalItems"`
	Items      []GBooksResponseItems `json:"items"`
}

type GBooksResponseItems struct {
	VolumeInfo GBooksResponseItemVolumeInfo `json:"volumeInfo"`
	SaleInfo   GBooksResponseItemSaleInfo   `json:"saleInfo"`
	AccessInfo GBooksResponseItemAccessInfo `json:"accessInfo"`
}

type GBooksResponseItemVolumeInfo struct {
	Title         string   `json:"title"`
	Subtitle      string   `json:"subtitle"`
	Authors       []string `json:"authors"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"publishedDate"`
	Description   string   `json:"description"`
	PageCount     int      `json:"pageCount"`
	Categories    []string `json:"categories"`
	Language      string   `json:"language"`
}

type GBooksResponseItemSaleInfo struct {
	IsEbook bool   `json:"isEbook"`
	BuyLink string `json:"buyLink"`
}

type GBooksResponseItemAccessInfo struct {
	PublicDomain bool `json:"publicDomain"`
}

func booksHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" && req.Body != nil {
		acceptHeaderValues := strings.Split(req.Header.Get("Accept"), ",")

		requestAcceptsJSONAsResponse := false
		for _, value := range acceptHeaderValues {
			value = strings.Trim(value, " ")
			if value == "application/json" || value == "*/*" {
				requestAcceptsJSONAsResponse = true
			}
		}

		if !requestAcceptsJSONAsResponse {
			message := `Header Accepts with value application/json not found in request`
			log.Printf(message)
			server.WriteErrorResp(415, message, rw)
			return
		}

		qParams := req.URL.Query()

		if qParams.Get("tittle") == "" {
			server.WriteErrorResp(404, "Filter tittle not present", rw)
			return
		}

		startIndex, err := strconv.Atoi(qParams.Get("startIndex"))
		if err != nil {
			startIndex = 0
		}

		filters := BookFilters{
			Language:   qParams.Get("lang"),
			Tittle:     qParams.Get("tittle"),
			Author:     qParams.Get("author"),
			Subject:    qParams.Get("subject"),
			StartIndex: startIndex,
		}

		books, err := findBooks(filters)
		if err != nil {
			message := "Could not retrieve books: " + err.Error()
			server.WriteErrorResp(500, message, rw)
			return
		}
		server.WriteRespWithStatus(rw, 200, books)
	} else {
		server.WriteErrorResp(404, "Not found", rw)
	}

}

func findBooks(filters BookFilters) ([]Book, error) {

	qParamFilters := getQParams(filters)
	u, _ := url.Parse(GoogleBooksBaseUrl + "/v1/volumes?" + qParamFilters)

	log.Printf("Makin request to %s ", u)
	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Host", "no-cache")
	if err != nil {
		log.Println("Error while trying to make a request to Google Books API. ", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("error while trying to make a request to Google Books API. Status code %d", resp.StatusCode)
		log.Println(err)

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		log.Println(buf.String())

		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var gBooksResp GBooksResponse
	err = decoder.Decode(&gBooksResp)
	if err != nil {
		log.Println("Error while trying to parse Google Books API response: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	return gBooksRespToBooks(gBooksResp), nil
}

func getQParams(filters BookFilters) (qParams string) {

	qParams += "q="

	if filters.Author != "" {
		qParams += fmt.Sprintf("inauthor:%s+", filters.Author)
	}

	if filters.Tittle != "" {
		qParams += fmt.Sprintf("intitle:%s+", filters.Tittle)
	}

	if filters.Subject != "" {
		qParams += fmt.Sprintf("subject:%s+", filters.Subject)
	}

	qParams = strings.TrimRight(qParams, "+")

	if filters.Language != "" {
		qParams += fmt.Sprintf("&langRestrict=%s", filters.Language)
	}

	startIndex := 0
	if filters.StartIndex > 1 {
		filters.StartIndex = 0
	}
	qParams += fmt.Sprintf("&startIndex=%d", startIndex)

	return strings.ReplaceAll(qParams, " ", "%20")
}

func gBooksRespToBooks(response GBooksResponse) []Book {

	books := make([]Book, len(response.Items))

	for i, item := range response.Items {
		b := Book{
			Language:      item.VolumeInfo.Language,
			Tittle:        item.VolumeInfo.Title,
			Authors:       item.VolumeInfo.Authors,
			Subjects:      item.VolumeInfo.Categories,
			Publisher:     item.VolumeInfo.Publisher,
			PublishedDate: item.VolumeInfo.PublishedDate,
			NumberOfPages: item.VolumeInfo.PageCount,
			Description:   item.VolumeInfo.Description,
			Ebook:         item.SaleInfo.IsEbook,
			PublicDomain:  item.AccessInfo.PublicDomain,
			LinkToBuy:     item.SaleInfo.BuyLink,
		}
		books[i] = b
	}
	return books
}
