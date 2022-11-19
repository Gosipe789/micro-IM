package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func NewMysql() (*sql.DB, error) {

	return sql.Open("mysql", "root:123456@/usdt")

	//queries := tutorial.New(db)
	//
	//// list all authors
	//authors, err := queries.ListAuthors(ctx)
	//if err != nil {
	//	return err
	//}
	//log.Println(authors)
	//
	//// create an author
	//result, err := queries.CreateAuthor(ctx, tutorial.CreateAuthorParams{
	//	Name: "Brian Kernighan",
	//	Bio:  sql.NullString{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
	//})
	//if err != nil {
	//	return err
	//}
	//
	//insertedAuthorID, err := result.LastInsertId()
	//if err != nil {
	//	return err
	//}
	//log.Println(insertedAuthorID)
	//
	//// get the author we just inserted
	//fetchedAuthor, err := queries.GetAuthor(ctx, insertedAuthorID)
	//if err != nil {
	//	return err
	//}
	//
	//// prints true
	//log.Println(reflect.DeepEqual(insertedAuthorID, fetchedAuthor.ID))
}
