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

type LessonHandler struct {
	Pool *pgxpool.Pool
}

func (h *LessonHandler) ListLessons(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var lessons []types.Lesson
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			"select lesson_id, lesson_name, chapter_id, provider_id from lessons")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l types.Lesson
			if err := rows.Scan(&l.Lesson_id, &l.Lesson_name, &l.Chapter_id, &l.Provider_id); err != nil {
				return err
			}
			lessons = append(lessons, l)
		}
		return rows.Err()
	})

	if err != nil {
		log.Println(err)
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		return
	}
	response.JSON(http.StatusOK, lessons)
}

func (h *LessonHandler) CreateLesson(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var lesson types.Lesson

	//Decoding
	if err := json.NewDecoder(request.Body()).Decode(&lesson); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	//Data validation
	if lesson.Lesson_name == "" || lesson.Chapter_id == 0 {
		response.Status(http.StatusBadRequest)
		response.Error("Lesson name and chapter id are required")
		return
	}

	//Insert into db
	var lessonID int
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into lessons (lesson_name, chapter_id, provider_id)
		values ($1, $2, $3)
		returning lesson_id`,
			lesson.Lesson_name, lesson.Chapter_id, provider_id,
		).Scan(&lessonID)
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error(fmt.Sprintf("db query error: %s", err.Error()))
		return
	}
	lesson.Lesson_id = lessonID
	lesson.Provider_id = provider_id
	response.JSON(http.StatusCreated, lesson)
}

func (h *LessonHandler) ShowLesson(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	lesson_id_str := request.RouteParams["lesson_id"]
	lesson_id, err := strconv.Atoi(lesson_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty lesson_id")
		return
	}

	var l types.Lesson
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"select lesson_id, lesson_name, chapter_id, provider_id from lessons where lesson_id = $1",
			lesson_id,
		).Scan(&l.Lesson_id, &l.Lesson_name, &l.Chapter_id, &l.Provider_id)
	})

	if err == pgx.ErrNoRows {
		response.Status(http.StatusNotFound)
		response.Error("No lesson found with this lesson_id")
		return
	} else if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("db query error")
		return
	}
	response.JSON(http.StatusOK, l)
}

func (h *LessonHandler) UpdateLesson(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	lesson_id_str := request.RouteParams["lesson_id"]
	lesson_id, err := strconv.Atoi(lesson_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty lesson_id")
		return
	}

	var lesson types.Lesson

	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&lesson); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	// Data validation
	if lesson.Lesson_name == "" || lesson.Chapter_id == 0 {
		response.Status(http.StatusBadRequest)
		response.Error("Lesson name and chapter id are needed")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx,
			`update lessons set lesson_name = $1, chapter_id = $2
		where lesson_id = $3`,
			lesson.Lesson_name, lesson.Chapter_id, lesson_id,
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
		response.Error("No lesson found for this provider")
		return
	}
	lesson.Provider_id = provider_id
	lesson.Lesson_id = lesson_id
	response.JSON(http.StatusOK, lesson)
}

func (h *LessonHandler) DeleteLesson(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	lesson_id_str := request.RouteParams["lesson_id"]
	lesson_id, err := strconv.Atoi(lesson_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty lesson_id")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx, "delete from lessons where lesson_id = $1", lesson_id)
		return e
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("Query to delete failed")
		return
	}

	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No lesson found with this lesson id")
		return
	}

	response.JSON(http.StatusOK, map[string]string{"deleted lesson id": lesson_id_str})
}
