package types

type Provider struct {
	Provider_id  int    `json:"provider_id"`
	Last_name    string `json:"last_name"`
	First_name   string `json:"first_name"`
	Phone_number string `json:"phone_number"`
	Email        string `json:"email"`
}

type Course struct {
	Course_id          int    `json:"course_id"`
	Course_name        string `json:"course_name"`
	Course_description string `json:"course_description"`
	Provider_id        int    `json:"provider_id"`
}

type Chapter struct {
	Chapter_id        int    `json:"chapter_id"`
	Chapter_name      string `json:"chapter_name"`
	Chapter_num       int    `json:"chapter_num"`
	Parent_chapter_id *int64 `json:"parent_chapter_id"`
	Course_id         int    `json:"course_id"`
	Provider_id       int    `json:"provider_id"`
	Level             int    `json:"level"`
}

type Lesson struct {
	Lesson_id   int    `json:"lesson_id"`
	Lesson_name string `json:"lesson_name"`
	Chapter_id  int    `json:"chapter_id"`
	Provider_id int    `json:"provider_id"`
}

type VideoMetadata struct {
	Video_id          int    `json:"video_id"`
	Title             string `json:"title"`
	Video_description string `json:"video_description"`
	Video_file_id     int    `json:"video_file_id"`
	Subtitle_text     string `json:"subtitle_text"`
	Lesson_id         int    `json:"lesson_id"`
	Provider_id       int    `json:"provider_id"`
}

// https://medium.com/@suleymanif.tural/sql-server-full-text-searching-practical-approach-50001151e475
// https://vela.simplyblock.io/blog/row-level-security-postgres/
// https://neon.com/postgresql/tutorial/recursive-query
// https://iniakunhuda.medium.com/postgresql-full-text-search-a-powerful-alternative-to-elasticsearch-for-small-to-medium-d9524e001fe0
// https://github.com/simplyblock/example-rls-invoicing/tree/main
