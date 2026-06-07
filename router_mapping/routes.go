package router_mapping

import (
	"tralgo/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
	"goyave.dev/goyave/v5"
)

func CreateRoutes(pool *pgxpool.Pool) func(
	server *goyave.Server, router *goyave.Router,
) {
	return func(server *goyave.Server, router *goyave.Router) {
		ph := &handlers.ProviderHandler{Pool: pool}
		router.Get("/providers", ph.ListProviders)
		router.Post("/providers", ph.CreateProvider)
		providersSubrouter := router.Subrouter("/providers/{provider_id}")
		providersSubrouter.Get("/", ph.ShowProvider)
		providersSubrouter.Put("/", ph.UpdateProvider)
		providersSubrouter.Delete("/", ph.DeleteProvider)

		ch := &handlers.CourseHandler{Pool: pool}
		coursesSubrouter := router.Subrouter("/courses")
		coursesSubrouter.Post("/", ch.CreateCourse)
		coursesSubrouter.Get("/", ch.ListCourses)
		coursesSubrouter.Get("/{course_id}", ch.ShowCourse)
		coursesSubrouter.Put("/{course_id}", ch.UpdateCourse)
		coursesSubrouter.Delete("/{course_id}", ch.DeleteCourse)

		chapter_h := &handlers.ChapterHandler{Pool: pool}
		chaptersSubrouter := router.Subrouter("/chapters")
		chaptersSubrouter.Post("/", chapter_h.CreateChapter)
		chaptersSubrouter.Get("/", chapter_h.ListChapters)
		chaptersSubrouter.Get("/{chapter_id}", chapter_h.ShowChapter)
		chaptersSubrouter.Put("/{chapter_id}", chapter_h.UpdateChapter)
		chaptersSubrouter.Delete("/{chapter_id}", chapter_h.DeleteChapter)

		lh := &handlers.LessonHandler{Pool: pool}
		lessonsSubrouter := router.Subrouter("/lessons")
		lessonsSubrouter.Post("/", lh.CreateLesson)
		lessonsSubrouter.Get("/", lh.ListLessons)
		lessonsSubrouter.Get("/{lesson_id}", lh.ShowLesson)
		lessonsSubrouter.Put("/{lesson_id}", lh.UpdateLesson)
		lessonsSubrouter.Delete("/{lesson_id}", lh.DeleteLesson)

		vh := &handlers.VideoHandler{Pool: pool}
		videosSubrouter := router.Subrouter("/videos")
		videosSubrouter.Post("/", vh.CreateVideo)
		videosSubrouter.Get("/", vh.ListVideos)
		videosSubrouter.Get("/{video_id}", vh.ShowVideo)
		videosSubrouter.Put("/{video_id}", vh.UpdateVideo)
		videosSubrouter.Delete("/{video_id}", vh.DeleteVideo)
	}
}
