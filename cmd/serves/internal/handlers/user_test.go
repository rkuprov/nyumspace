package handlers

import (
	"github.com/rkuprov/nyumspace/pkg/tests"
	"log"
	"testing"
	"time"
)

func TestServerHandler_RegisterUser(t *testing.T) {
	pool := tests.DBForTest(t)
	time.After(1 * time.Second)

	log.Printf(`Database created: %s`, pool.Config().ConnConfig.Database)

	defer func() {
		name := pool.Config().ConnConfig.Database
		err := tests.RemoveDBForTest(pool)
		if err != nil {
			log.Fatalf("failed to remove test database: %v", err)
		}
		log.Printf("Database %s removed", name)
	}()
	defer pool.Close()
}
