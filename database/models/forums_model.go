package models

import (
	"bytes"

	"github.com/jackc/pgx"

	"strconv"

	"../../utils"
)

type Forums struct {
	FID     int64  `json:"fid"`
	Posts   int64  `json:"posts"`
	Slug    string `json:"slug"`
	Threads int32  `json:"threads"`
	Title   string `json:"title"`
	Author  string `json:"user"`
}

func (forum *Forums) CreateForum(pool *pgx.ConnPool) error {
	var id int64
	err := pool.QueryRow(`INSERT INTO forums(slug, title, author)`+
		`VALUES ($1, $2, $3) RETURNING "fID";`,
		forum.Slug, forum.Title, forum.Author).Scan(&id)
	if err != nil {
		// if pgerr, ok := err.(pgx.PgError); ok {
		// 	if pgerr.ConstraintName == "index_on_forums_slug" {
				return utils.UniqueError
		// 	} else {
		// 		return err
		// 	}
		// }
		// return err
	}
	return nil
}

func (forum *Forums) GetForumBySlug(pool *pgx.ConnPool) error {
	return pool.QueryRow(`SELECT slug, title, author, posts, threads  FROM forums WHERE slug = $1`,
		forum.Slug).Scan(&forum.Slug, &forum.Title, &forum.Author, &forum.Posts, &forum.Threads)
}

func (forum *Forums) GetAllThreads(pool *pgx.ConnPool, limit, since, desc string) ([]Threads, error) {
	queryRow := bytes.Buffer{}
	queryRow.WriteString(`SELECT "tID", author, created, forum, message, slug, title, votes FROM threads WHERE forum = $1`)

	var params []interface{}
	params = append(params, forum.Slug)
	if since != "" {
		if desc == "true" {
			queryRow.WriteString(` AND created <= $`)
			queryRow.WriteString(strconv.Itoa(len(params) + 1))
		} else {
			queryRow.WriteString(` AND created >= $`)
			queryRow.WriteString(strconv.Itoa(len(params) + 1))
		}
		params = append(params, since)
	}
	if desc == "true" {
		queryRow.WriteString(` ORDER BY created DESC`)
	} else {
		queryRow.WriteString(` ORDER BY created ASC`)
	}
	if limit != "" {
		queryRow.WriteString(` LIMIT $`)
		queryRow.WriteString(strconv.Itoa(len(params) + 1))
		params = append(params, limit)
	}

	rows, err := pool.Query(queryRow.String(), params...)
	if err != nil {
		return nil, err
	}

	resultThreads := []Threads{}

	currentThreadInRows := Threads{}
	for rows.Next() {
		rows.Scan(&currentThreadInRows.TID, &currentThreadInRows.Author, &currentThreadInRows.Created, &currentThreadInRows.Forum,
			&currentThreadInRows.Message, &currentThreadInRows.Slug, &currentThreadInRows.Title, &currentThreadInRows.Votes)
		resultThreads = append(resultThreads, currentThreadInRows)
	}
	return resultThreads, nil
}

func (forum *Forums) GetMembers(pool *pgx.ConnPool, limit, since, desc string) ([]Users, error) {
	queryRow := bytes.Buffer{}
	queryRow.WriteString(`SELECT u.about, u.email, u.fullname, u.nickname FROM members AS m
 	JOIN users as u ON u.nickname = m.author AND m.forum = $1`)

	var params []interface{}
	params = append(params, forum.Slug)
	if since != "" {
		if desc == "true" {
			queryRow.WriteString(` AND u.nickname < $`)
			queryRow.WriteString(strconv.Itoa(len(params) + 1))
		} else {
			queryRow.WriteString(` AND u.nickname > $`)
			queryRow.WriteString(strconv.Itoa(len(params) + 1))
		}
		params = append(params, since)
	}
	if desc == "true" {
		queryRow.WriteString(` ORDER BY u.nickname DESC`)
	} else {
		queryRow.WriteString(` ORDER BY u.nickname ASC`)
	}
	if limit != "" {
		queryRow.WriteString(` LIMIT $`)
		queryRow.WriteString(strconv.Itoa(len(params) + 1))
		params = append(params, limit)
	}

	rows, err := pool.Query(queryRow.String(), params...)
	if err != nil {
		return nil, err
	}

	resultUsers := []Users{}

	currentUserInRows := Users{}
	for rows.Next() {
		rows.Scan(&currentUserInRows.About, &currentUserInRows.Email, &currentUserInRows.Fullname, &currentUserInRows.Nickname)
		resultUsers = append(resultUsers, currentUserInRows)
	}
	return resultUsers, nil
}

func ForumsCount(pool *pgx.ConnPool) (int32, error) {
	var count int32
	err := pool.QueryRow("SELECT COUNT(*) FROM forums").Scan(&count)
	return count, err
}
