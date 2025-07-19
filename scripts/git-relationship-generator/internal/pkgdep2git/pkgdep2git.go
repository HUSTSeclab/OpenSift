package pkgdep2git
import (
    "database/sql"
	"log"
	"strings"
	"fmt"
)
var repoList = []string{
	"arch",
	"debian",
	"gentoo",
	"nix",
	"homebrew",
}
func FetchAlldep(db *sql.DB) map[string]map[[2]string]struct{} {
	var repodepSet = make(map[string]map[[2]string]struct{})

	for _, repo := range repoList {
		rows, err := db.Query("SELECT frompackage, topackage FROM " + repo + "_relationships")
		if err != nil {
			log.Println("Error querying " + repo + "_relationships:", err)
			log.Fatal(err)	
		}
		defer rows.Close()

		if _, exists := repodepSet[repo]; !exists {
			repodepSet[repo] = make(map[[2]string]struct{})
		}

		for rows.Next() {
			var frompackage sql.NullString
			var topackage sql.NullString

			if err := rows.Scan(&frompackage, &topackage); err != nil {
				log.Println("Error scanning row:", err)
				log.Fatal(err)
			}

			if frompackage.Valid && topackage.Valid {
				key := [2]string{frompackage.String, topackage.String}
				repodepSet[repo][key] = struct{}{}
			}
		}
	}
	return repodepSet
}


func GenGitDep(db *sql.DB, repodepMap map[string]map[[2]string]struct{}) map[[2]string]struct{} {
	gitdepMap := make(map[[2]string]struct{})

	for repo, depsMap := range repodepMap {
		for deps := range depsMap {
			frompkg := deps[0]
			topkg := deps[1]
			fromgit, togit := FetchPkg2Git(db, frompkg, topkg, repo)
			if fromgit != "" && togit != "" {
				key := [2]string{fromgit, togit}
				gitdepMap[key] = struct{}{}
			}
		}
	}
	return gitdepMap
}


func FetchPkg2Git(db *sql.DB, frompkg string, topkg string, repo string) (string, string) {
	var fromgit sql.NullString
	var togit sql.NullString

	queryFrom := fmt.Sprintf("SELECT git_link FROM %s WHERE package = $1", repo + "_packages")
	queryTo := fmt.Sprintf("SELECT git_link FROM %s WHERE package = $1", repo + "_packages")

	err := db.QueryRow(queryFrom, frompkg).Scan(&fromgit)
	if err != nil {
		log.Printf("No git_link found for package %s in repo %s, setting to empty, from\n", frompkg, repo)
		log.Fatal(err)
	}

	err = db.QueryRow(queryTo, topkg).Scan(&togit)
	if err == sql.ErrNoRows {
		togit.Valid = false
	} else if err != nil {
		log.Printf("No git_link found for package %s in repo %s, setting to empty, to\n", topkg, repo)
		log.Fatal(err)
	}

	if fromgit.Valid && togit.Valid {
		return fromgit.String, togit.String
	}
	return "", ""
}


func BatchUpdate(db *sql.DB, batchSize int, gitdepMap map[[2]string]struct{}) error {
	keys := make([][2]string, 0, len(gitdepMap))
	for key := range gitdepMap {
		keys = append(keys, key)
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]

		query := "INSERT INTO git_relationships (fromgitlink, togitlink) VALUES "
		values := []interface{}{}
		placeholders := []string{}

		for j, key := range batch {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", j*2+1, j*2+2))
			values = append(values, key[0], key[1])
		}

		query += strings.Join(placeholders, ", ")

		_, err := db.Exec(query, values...)
		if err != nil {
			return fmt.Errorf("error inserting batch: %w", err)
		}
	}
	return nil
}