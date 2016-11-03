package main

import (
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/olahol/melody"
  "net/http"
  "os"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "time"
  "encoding/json"
)

type Msg struct {
  Username, Content string
}

func main() {
  r := gin.Default()
  m := melody.New()

  db_url := os.Getenv("SHAREDCHAT_DB_URL")
  db, _ := sql.Open("mysql", db_url)

  r.GET("/", func(c *gin.Context) {
    http.ServeFile(c.Writer, c.Request, "./public/index.html")
  })

  r.GET("/allMessages", func(c *gin.Context) {
    var (
      content string
      allMessages []string
    )

    rows, _ := db.Query("SELECT content FROM messages ORDER BY timestamp ASC")
    defer rows.Close()
    for rows.Next() {
      err := rows.Scan(&content)
      if err != nil {
        fmt.Println(err)
      }
      allMessages = append(allMessages, content)
    }

    c.JSON(200, gin.H{
      "messages": allMessages,
    })
  })

  r.GET("/ws", func(c *gin.Context) {
    m.HandleRequest(c.Writer, c.Request)
  })

  m.HandleMessage(func(s *melody.Session, message []byte) {
    var msg_data Msg
    err := json.Unmarshal(message, &msg_data)
    if err != nil {
      fmt.Println(err)
    }

    insert_stmt, _ := db.Prepare("INSERT messages SET user_id=?,message_type=?,content=?,timestamp=?")
    _, err = insert_stmt.Exec(1, "text", msg_data.Content, time.Now())
    if err != nil {
      fmt.Println(err)
    }
    m.Broadcast([]byte(msg_data.Content))
  })

  r.Static("/public", "./public")
  r.Run(":5000")
}
