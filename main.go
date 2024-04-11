package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var urlExample = "postgres://postgres:13761380@localhost:5432/recordings"
var conn, err = pgx.Connect(context.Background(), urlExample)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Price  int    `json:"price"`
}

func getBooks(c *gin.Context) {
	var books []Book
	rows, err := conn.Query(context.Background(), "SELECT * FROM book")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query Failed. %v\n", err)
		os.Exit(1)
	}
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price); err != nil {
			return
		}
		books = append(books, b)
	}
	c.IndentedJSON(http.StatusOK, books)
}

func postBook(c *gin.Context) {
	var newBook Book

	if err := c.BindJSON(&newBook); err != nil {
		return
	}
	tx, err := conn.Begin(context.Background())
	if err != nil {
		fmt.Errorf("error %v", err)
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(context.Background(), fmt.Sprintf("INSERT INTO book (title, author, price) VALUES('%s', '%s', %d)", newBook.Title, newBook.Author, newBook.Price))
	if err != nil {
		fmt.Errorf("add book %v", err)
	}
	err = tx.Commit(context.Background())
	if err != nil {
		fmt.Printf("err %v", err)
	}

	c.IndentedJSON(http.StatusOK, newBook)

}

func getBookById(c *gin.Context) {
	var book Book
	id := c.Param("id")
	ans, _ := strconv.Atoi(id)
	row := conn.QueryRow(context.Background(), fmt.Sprintf("SELECT * FROM book WHERE id=%d", ans))
	if err := row.Scan(&book.ID, &book.Title, &book.Author, &book.Price); err != nil {
		fmt.Errorf("error")
	}
	c.IndentedJSON(http.StatusOK, book)
}

func deleteBookById(c *gin.Context) {
	id := c.Param("id")
	ans, _ := strconv.Atoi(id)
	commandTag, err := conn.Exec(context.Background(), fmt.Sprintf("DELETE FROM book WHERE id = %d", ans))

	if err != nil {
		fmt.Println("err")
	}
	if commandTag.RowsAffected() != 1 {
		 errors.New("No Row Found")
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"message": "The row was deleted.",
	})

}

func main() {
	// urlExample := "postgres://postgres:13761380@localhost:5432/recordings"
	// conn, err := pgx.Connect(context.Background(), urlExample)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable To connect to database %v\n", err)
	// 	os.Exit(1)
	// }
	defer conn.Close(context.Background())

	router := gin.Default()
	router.GET("/", getBooks)
	router.POST("/", postBook)
	router.GET("/:id", getBookById)
	router.DELETE("/:id", deleteBookById)
	router.Run("localhost:8080")

}
