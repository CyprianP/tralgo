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

type CourseHandler struct {
	Pool *pgxpool.Pool
}

func (h *CourseHandler) ListCourses(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var courses []types.Course
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			"select course_id, course_name, course_description, provider_id from courses")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var c types.Course
			if err := rows.Scan(&c.Course_id, &c.Course_name, &c.Course_description, &c.Provider_id); err != nil {
				return err
			}
			courses = append(courses, c)
		}
		return rows.Err()
	})

	if err != nil {
		log.Println(err)
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		return
	}
	response.JSON(http.StatusOK, courses)
}

func (h *CourseHandler) CreateCourse(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var course types.Course

	//Decoding
	if err := json.NewDecoder(request.Body()).Decode(&course); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	//Data validation
	if course.Course_name == "" || course.Course_description == "" {
		response.Status(http.StatusBadRequest)
		response.Error("Course name and Course description are required")
		return
	}

	//Insert into db
	var courseID int
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into courses (course_name, course_description, provider_id)
		values ($1, $2, $3)
		returning course_id`,
			course.Course_name, course.Course_description, provider_id,
		).Scan(&courseID)
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error(fmt.Sprintf("db query error: %s", err.Error()))
		return
	}
	course.Course_id = courseID
	course.Provider_id = provider_id
	response.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) ShowCourse(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	course_id_str := request.RouteParams["course_id"]
	course_id, err := strconv.Atoi(course_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty course_id")
		return
	}

	var c types.Course
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"select course_id, course_name, course_description, provider_id from courses where course_id = $1",
			course_id,
		).Scan(&c.Course_id, &c.Course_name, &c.Course_description, &c.Provider_id)
	})

	if err == pgx.ErrNoRows {
		response.Status(http.StatusNotFound)
		response.Error("No course found with this ID")
		return
	} else if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("db query error")
		return
	}
	response.JSON(http.StatusOK, c)
}

func (h *CourseHandler) UpdateCourse(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	course_id_str := request.RouteParams["course_id"]
	course_id, err := strconv.Atoi(course_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty course_id")
		return
	}

	var course types.Course

	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&course); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	// Data validation
	if course.Course_name == "" || course.Course_description == "" {
		response.Status(http.StatusBadRequest)
		response.Error("Course name and Course description are needed")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx,
			`update courses set course_name = $1, course_description = $2
		where course_id = $3`,
			course.Course_name, course.Course_description, course_id,
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
		response.Error("No course found for this provider")
		return
	}
	course.Provider_id = provider_id
	course.Course_id = course_id
	response.JSON(http.StatusOK, course)
}

func (h *CourseHandler) DeleteCourse(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	course_id_str := request.RouteParams["course_id"]
	course_id, err := strconv.Atoi(course_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty course_id")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx, "delete from courses where course_id = $1", course_id)
		return e
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("Query to delete failed")
		return
	}

	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No course found with this course id")
		return
	}

	response.JSON(http.StatusOK, map[string]string{"deleted course id": course_id_str})
}
