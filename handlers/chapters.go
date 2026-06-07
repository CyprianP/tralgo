package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"tralgo/tenantized"
	"tralgo/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"goyave.dev/goyave/v5"
)

type ChapterHandler struct {
	Pool *pgxpool.Pool
}

func (h *ChapterHandler) ListChapters(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var chapters []types.Chapter
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`
		with recursive sub_chapters as (
			select 
				chapter_id, chapter_name, chapter_num, parent_chapter_id, course_id, provider_id,
				1 as level
			from 
				chapters
			where 
				parent_chapter_id is null
			
			union all

			-- recursive part to find all subchapters
			select
				c.chapter_id, c.chapter_name, c.chapter_num, c.parent_chapter_id, c.course_id, c.provider_id,
				s.level + 1
			from 
				chapters c
			inner join sub_chapters s
				on c.parent_chapter_id = s.chapter_id
			)
		select 
			*
		from 
			sub_chapters
		order by
			course_id, level, chapter_id
		`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c types.Chapter
			if err := rows.Scan(&c.Chapter_id, &c.Chapter_name, &c.Chapter_num, &c.Parent_chapter_id, &c.Course_id, &c.Provider_id, &c.Level); err != nil {
				return err
			}
			chapters = append(chapters, c)
		}
		return rows.Err()
	})

	if err != nil {
		log.Println(err)
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		return
	}
	response.JSON(http.StatusOK, chapters)
}

func (h *ChapterHandler) CreateChapter(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var chapter types.Chapter

	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&chapter); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	// Data validation
	if chapter.Chapter_name == "" || chapter.Course_id == 0 {
		response.Status(http.StatusBadRequest)
		response.Error("Chapter name and course id are required")
		return
	}

	// Insert into db
	var chapterID int
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into chapters (chapter_name, chapter_num, parent_chapter_id, course_id, provider_id)
		values ($1, $2, $3, $4, $5)
		returning chapter_id`,
			chapter.Chapter_name, chapter.Chapter_num, chapter.Parent_chapter_id, chapter.Course_id, provider_id,
		).Scan(&chapterID)
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error(fmt.Sprintf("db query error: %s", err.Error()))
		return
	}
	chapter.Chapter_id = chapterID
	chapter.Provider_id = provider_id
	response.JSON(http.StatusCreated, chapter)
}

func (h *ChapterHandler) ShowChapter(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	chapter_id_str := request.RouteParams["chapter_id"]
	chapter_id, err := strconv.Atoi(chapter_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty chapter_id")
		return
	}

	var c types.Chapter
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"select chapter_id, chapter_name, chapter_num, parent_chapter_id, course_id, provider_id from chapters where chapter_id = $1",
			chapter_id,
		).Scan(&c.Chapter_id, &c.Chapter_name, &c.Chapter_num, &c.Parent_chapter_id, &c.Course_id, &c.Provider_id)
	})

	if err == pgx.ErrNoRows {
		response.Status(http.StatusNotFound)
		response.Error("No chapter found with this chapter_id")
		return
	} else if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("db query error")
		return
	}
	response.JSON(http.StatusOK, c)
}

func (h *ChapterHandler) UpdateChapter(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	chapter_id_str := request.RouteParams["chapter_id"]
	chapter_id, err := strconv.Atoi(chapter_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty chapter_id")
		return
	}

	var chapter types.Chapter

	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&chapter); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	// Data validation
	if chapter.Chapter_name == "" {
		response.Status(http.StatusBadRequest)
		response.Error("Chapter name is needed")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx,
			`update chapters set chapter_name = $1, chapter_num = $2, parent_chapter_id = $3
		where chapter_id = $4`,
			chapter.Chapter_name, chapter.Chapter_num, chapter.Parent_chapter_id, chapter_id,
		)
		return e
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		log.Println(err)
		return
	}

	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No chapter found for this provider")
		return
	}

	chapter.Provider_id = provider_id
	chapter.Chapter_id = chapter_id
	response.JSON(http.StatusOK, chapter)
}

func (h *ChapterHandler) DeleteChapter(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	chapter_id_str := request.RouteParams["chapter_id"]
	chapter_id, err := strconv.Atoi(chapter_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty chapter_id")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx, "delete from chapters where chapter_id = $1", chapter_id)
		return e
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("Query to delete failed")
		return
	}

	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No chapter found with this chapter id")
		return
	}

	response.JSON(http.StatusOK, map[string]string{"deleted chapter id": chapter_id_str})
}
