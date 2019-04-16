package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
)

var db *sql.DB

//定义person类型结构
type Person struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

//定义一个getALL函数用于回去全部的信息
func (p Person) getAll() (persons []Person, err error) {
	rows, err := db.Query("SELECT id, first_name, last_name FROM person")
	if err != nil {
		return
	}
	for rows.Next() {
		var person Person
		//遍历表中所有行的信息
		rows.Scan(&person.Id, &person.FirstName, &person.LastName)
		//将person添加到persons中
		persons = append(persons, person)
	}
	//最后关闭连接
	defer rows.Close()
	return
}

//通过id查询
func (p Person) get() (person Person, err error) {
	row := db.QueryRow("SELECT id, first_name, last_name FROM person WHERE id=?", p.Id)
	err = row.Scan(&person.Id, &person.FirstName, &person.LastName)
	if err != nil {
		return
	}
	return
}
func (p Person) add() (Id int, err error) {
	stmt, err := db.Prepare("INSERT INTO person(first_name, last_name) VALUES (?, ?)")
	if err != nil {
		return
	}
	//执行插入操作
	rs, err := stmt.Exec(p.FirstName, p.LastName)
	if err != nil {
		return
	}
	//返回插入的id
	id, err := rs.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	//将id类型转换
	Id = int(id)
	defer stmt.Close()
	return
}
//通过id删除
func (p Person) del() (rows int, err error) {
	stmt, err := db.Prepare("DELETE FROM person WHERE id=?")
	if err != nil {
		log.Fatalln(err)
	}

	rs, err := stmt.Exec(p.Id)
	if err != nil {
		log.Fatalln(err)
	}
	//删除的行数
	row, err := rs.RowsAffected()
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	rows = int(row)
	return
}

func main() {

	var err error
	db, err = sql.Open("mysql", "root:UvcO12345@@tcp(127.0.0.1:3306)/springcloudweb?parseTime=true")
	//错误检查
	if err != nil {
		log.Fatal(err.Error())
	}
	//推迟数据库连接的关闭
	defer db.Close()

	//
	err = db.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}
	//创建一个路由Handler
	router := gin.Default()

	//get方法的查询
	router.GET("/person", func(c *gin.Context) {
		p := Person{}
		persons, err := p.getAll()
		if err != nil {
			log.Fatal(err)
		}
		//H is a shortcut for map[string]interface{}
		c.JSON(http.StatusOK, gin.H{
			"result": persons,
			"count":  len(persons),
		})
	})
	//利用get方法通过id查询
	router.GET("/person/:id", func(c *gin.Context) {
		var result gin.H
		//c.Params方法可以获取到/person/:id中的id值
		id := c.Param("id")
		Id, err := strconv.Atoi(id)
		if err != nil {
			log.Fatal(err)
		}
		//定义person结构
		p := Person{
			Id: Id,
		}
		person, err := p.get()
		if err != nil {
			result = gin.H{
				"result": nil,
				"count":  0,
			}
		} else {
			result = gin.H{
				"result": person,
				"count":  1,
			}
		}
		c.JSON(http.StatusOK, result)

	})
	//利用post方法新增数据
	router.POST("/person", func(c *gin.Context) {
		var p Person
		err := c.Bind(&p)
		if err != nil {
			log.Fatal(err)
		}
		Id, err := p.add()
		fmt.Print("id=", Id)
		name := p.FirstName + " " + p.LastName
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%s 插入成功", name),
		})
	})
	//利用DELETE请求方法通过id删除
	router.DELETE("/person/:id", func(c *gin.Context) {
		id := c.Param("id")

		Id, err := strconv.ParseInt(id, 10, 10)
		if err != nil {
			log.Fatalln(err)
		}
		p := Person{Id: int(Id)}
		rows, err := p.del()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("delete rows ", rows)

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully deleted user: %s", id),
		})
	})
	router.Run(":8080")
}