package main

import (
	"flag"
	"log"
	"prism/prism/cmd"
	"prism/prism/schema/migrations"
)

func main() {
	dbUri := flag.String("db", "", "database uri")
	version := flag.String("version", "", "version to migrate to")
	rollback := flag.Bool("rollback", false, "if the desired migration is a rollback")
	flag.Parse()

	if *dbUri == "" {
		log.Fatal("must specify --db arg")
	}

	if *version == "" {
		log.Fatal("must specify --version arg")
	}

	db := cmd.OpenDB(*dbUri)

	migrator := migrations.GetMigrator(db)

	if *rollback {
		log.Println("rolling back to version", *version)
		if err := migrator.RollbackTo(*version); err != nil {
			log.Fatalf("db rollback failed: %v", err)
		}
		log.Println("db rollback complete")
	} else {
		log.Println("migrating to version", *version)
		if err := migrator.MigrateTo(*version); err != nil {
			log.Fatalf("db migration failed: %v", err)
		}
		log.Println("db migration complete")
	}
}
