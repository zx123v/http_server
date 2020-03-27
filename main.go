package main

import (
    "os"
	"fmt"
    "time"
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/mysql"    
    "golang.org/x/crypto/bcrypt"
    "github.com/dgrijalva/jwt-go"
)

// db 設定
const (
	username = "root"
	password = "root"
	host     = "127.0.0.1"
	port     = "3306"
	dbname   = "go"
)
var db *gorm.DB

// accounts表結構
type Account struct {
    ID          int    `gorm:"primary_key"`
    Username    string `type:varchar(20); not null;"`
    Password    string `type:varchar(40); not null;`
    Mail        string `type:varchar(40); not null;`
    Token       string `type:varchar(256);null;`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   time.Time
}

// JWT 結構
type Token struct {
    UserId uint
    jwt.StandardClaims
}

// accounts model
type Accounts struct {
    gorm.Model
	Username   string `form:"Username" json:"Username" xml:"Username" binding:"required"`
	Password   string `form:"Password" json:"Password" xml:"Password" binding:"required"`
	Mail       string `form:"Mail" json:"Mail" xml:"Mail" binding:"required"`
    Token      string `json:"Token";sql:"-"`
}

type Login struct {
	Mail     string `form:"Mail" json:"Mail" xml:"Mail"  binding:"required"`
	Password string `form:"Password" json:"Password" xml:"Password" binding:"required"`
}

func main() {
	initDB()

	router := gin.Default()

	router.LoadHTMLGlob("views/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	router.GET("/hello", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("Hello, It Home!"))
	})

	v1 := router.Group("/v1")
	{
		v1.POST("/login", loginV1)
		v1.POST("/register", registerV1)
	}
	
	router.Run(":8888")
}

func initDB() {
    var err error
    path := strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", dbname, "?charset=utf8&parseTime=True&loc=Local"}, "")
    db, err = gorm.Open("mysql", path)
    if err != nil {
        panic(err)
    }
    db.DB().SetMaxIdleConns(10)
    db.DB().SetMaxOpenConns(100)

    if !db.HasTable(&Account{}) {
        if err := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&Account{}).Error; err != nil {
            panic(err)
        }
    }

    fmt.Println("connnect success")
}

func registerV1(c *gin.Context) {
    var data Accounts

    if err := c.ShouldBind(&data); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    temp := &Accounts{}
    err := db.Table("accounts").Where("Mail = ?", data.Mail).First(temp).Error

    if err != nil && err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusOK, gin.H{
            "status":  0,
            "message": "連接失敗",
        })
        return
    }

    if temp.ID > 0 {
        c.JSON(http.StatusOK, gin.H{
            "status":  0,
            "message": "創建失敗，用戶已存在!",
        })
        return
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)

    data.Password = string(hashedPassword)
    db.Create(&data)

    if data.ID <= 0 {
        c.JSON(http.StatusOK, gin.H{
            "status":  0,
            "message": "創建失敗",
        })
        return 
    }

    tk := &Token{UserId: data.ID}
    token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
    tokenString, _ := token.SignedString([]byte(os.Getenv("TOKEN_PASSWORD")))
    data.Token = tokenString
    db.Save(&data)

	c.JSON(http.StatusOK, gin.H{
		"status":  1,
        "token": data.Token,
		"message": "創建成功",
	})
}

func loginV1(c *gin.Context) {
	var data Login
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    temp := &Accounts{}
    err := db.Table("accounts").Where("Mail = ?", data.Mail).First(temp).Error

    if err != nil && err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusOK, gin.H{
            "status":  0,
            "message": "連接失敗",
        })
        return
    }

    if temp.ID <= 0 {
        c.JSON(http.StatusOK, gin.H{
            "status":  0,
            "message": "登入失敗，用戶不存在!",
        })
        return
    }

    //密碼驗證
    err = bcrypt.CompareHashAndPassword([]byte(temp.Password), []byte(data.Password))
    if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { 
        c.JSON(http.StatusOK, gin.H{
            "status":  0,
            "temp": temp.Password,
            "data": data.Password,
            "message": "登入失敗，密碼錯誤!",
        })
        return
    }   
	
	c.JSON(http.StatusOK, gin.H{
        "status":  1,
        "token": temp.Token,
        "message": "登入成功",
    })
}

