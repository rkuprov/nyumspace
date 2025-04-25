package main

import (
	"context"
	"nyum/pkg/daemon"
)

func main() {
	daemon.Run(func(ctx context.Context, d daemon.Daemon) error {
		//type Post struct {
		//	ID    int
		//	Title string
		//	Body  string
		//}
		//var post Post
		//err := d.DB.QueryRow(ctx, "select * from post where id = 1").Scan(&post.ID, &post.Title, &post.Body)
		//if err != nil {
		//	return err
		//}
		//fmt.Println(post)

		panic("crazy crash")

		return nil
	})
}
