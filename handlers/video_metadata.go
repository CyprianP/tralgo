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

type VideoHandler struct {
	Pool *pgxpool.Pool
}

func (h *VideoHandler) ListVideos(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var videos []types.VideoMetadata
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			"select video_id, title, video_description, video_file_id, subtitle_text, lesson_id, provider_id from video_metadata")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var v types.VideoMetadata
			if err := rows.Scan(&v.Video_id, &v.Title, &v.Video_description, &v.Video_file_id, &v.Subtitle_text, &v.Lesson_id, &v.Provider_id); err != nil {
				return err
			}
			videos = append(videos, v)
		}
		return rows.Err()
	})

	if err != nil {
		log.Println(err)
		response.Status(http.StatusInternalServerError)
		response.Error("db query failed")
		return
	}
	response.JSON(http.StatusOK, videos)
}

func (h *VideoHandler) CreateVideo(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	var video types.VideoMetadata

	//Decoding
	if err := json.NewDecoder(request.Body()).Decode(&video); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	//Data validation
	if video.Title == "" || video.Lesson_id == 0 {
		response.Status(http.StatusBadRequest)
		response.Error("Title and lesson id are required")
		return
	}

	//Insert into db
	var videoID int
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`insert into video_metadata (title, video_description, video_file_id, subtitle_text, lesson_id, provider_id)
		values ($1, $2, $3, $4, $5, $6)
		returning video_id`,
			video.Title, video.Video_description, video.Video_file_id, video.Subtitle_text, video.Lesson_id, provider_id,
		).Scan(&videoID)
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error(fmt.Sprintf("db query error: %s", err.Error()))
		return
	}
	video.Video_id = videoID
	video.Provider_id = provider_id
	response.JSON(http.StatusCreated, video)
}

func (h *VideoHandler) ShowVideo(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	video_id_str := request.RouteParams["video_id"]
	video_id, err := strconv.Atoi(video_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty video_id")
		return
	}

	var v types.VideoMetadata
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			"select video_id, title, video_description, video_file_id, subtitle_text, lesson_id, provider_id from video_metadata where video_id = $1",
			video_id,
		).Scan(&v.Video_id, &v.Title, &v.Video_description, &v.Video_file_id, &v.Subtitle_text, &v.Lesson_id, &v.Provider_id)
	})

	if err == pgx.ErrNoRows {
		response.Status(http.StatusNotFound)
		response.Error("No video found with this video_id")
		return
	} else if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("db query error")
		return
	}
	response.JSON(http.StatusOK, v)
}

func (h *VideoHandler) UpdateVideo(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	video_id_str := request.RouteParams["video_id"]
	video_id, err := strconv.Atoi(video_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty video_id")
		return
	}

	var video types.VideoMetadata

	// Decoding
	if err := json.NewDecoder(request.Body()).Decode(&video); err != nil {
		response.Status(http.StatusBadRequest)
		response.Error(fmt.Sprintf("invalid requestbody: %s", err.Error()))
		return
	}

	// Data validation
	if video.Title == "" || video.Lesson_id == 0 {
		response.Status(http.StatusBadRequest)
		response.Error("Title and lesson id are needed")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx,
			`update video_metadata set title = $1, video_description = $2, video_file_id = $3, subtitle_text = $4, lesson_id = $5
		where video_id = $6`,
			video.Title, video.Video_description, video.Video_file_id, video.Subtitle_text, video.Lesson_id, video_id,
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
		response.Error("No video found for this provider")
		return
	}
	video.Provider_id = provider_id
	video.Video_id = video_id
	response.JSON(http.StatusOK, video)
}

func (h *VideoHandler) DeleteVideo(response *goyave.Response, request *goyave.Request) {
	ctx := request.Context()

	provider_id, err := strconv.Atoi(request.Header().Get("X-Provider-ID"))
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("missing or invalid X-Provider-ID")
		return
	}

	video_id_str := request.RouteParams["video_id"]
	video_id, err := strconv.Atoi(video_id_str)
	if err != nil {
		response.Status(http.StatusBadRequest)
		response.Error("Invalid or empty video_id")
		return
	}

	var tag pgconn.CommandTag
	err = tenantized.WithTenant(ctx, h.Pool, provider_id, func(tx pgx.Tx) error {
		var e error
		tag, e = tx.Exec(ctx, "delete from video_metadata where video_id = $1", video_id)
		return e
	})

	if err != nil {
		response.Status(http.StatusInternalServerError)
		response.Error("Query to delete failed")
		return
	}

	if tag.RowsAffected() == 0 {
		response.Status(http.StatusNotFound)
		response.Error("No video found with this video id")
		return
	}

	response.JSON(http.StatusOK, map[string]string{"deleted video id": video_id_str})
}
