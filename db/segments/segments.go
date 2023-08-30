package segments

import (
	"database/sql"
	"time"
)

type SegmentRepository interface {
	CreateSegment(name string)
	DeleteSegment(name string)
	AddUserSegment(id int, segment string, tim string)
	ReturnSegment(id int) *sql.Rows
	DeleteUserSegment(id int, segment string)
	GetUserHistory(id int, startDate string, endDate string) *sql.Rows
	CheckTTL()
	DistributeUsers() *sql.Rows
	GetSegments() *sql.Rows
}

type Repository struct {
	repo SegmentRepository
	db   *sql.DB
}

func New(db *sql.DB, repo SegmentRepository) *Repository {
	return &Repository{
		repo: repo,
		db:   db,
	}
}

func (r *Repository) CreateSegment(name string) {

	_, err := r.db.Exec("call add_segment($1)", name)
	if err != nil {
		panic(err)
	}

}

func (r *Repository) DeleteSegment(name string) {
	_, err := r.db.Exec("call del_segment($1)", name)

	if err != nil {
		panic(err)
	}

}

func (r *Repository) CheckTTL() {
	ticker := time.NewTicker(1 * time.Hour)
	for _ = range ticker.C {
		_, err := r.db.Exec("call check_ttl($1)", time.Now())
		if err != nil {
			panic(err)
		}
	}
}

func (r *Repository) AddUserSegment(id int, segment string, tim string) {

	_, err := r.db.Exec("call add_user_seg($1,$2,$3)", id, segment, tim)
	if err != nil {
		panic(err)
	}
	_, err = r.db.Exec("call add_to_history($1,$2,$3,$4)", id, segment, time.Now(), "Added")
	if err != nil {
		panic(err)
	}

}

func (r *Repository) DeleteUserSegment(id int, segment string) {
	_, err := r.db.Exec("call del_user_seg($1,$2)", id, segment)
	if err != nil {
		panic(err)
	}
	_, err = r.db.Exec("call add_to_history($1,$2,$3,$4)", id, segment, time.Now(), "Deleted")
	if err != nil {
		panic(err)
	}

}

func (r *Repository) ReturnSegment(id int) *sql.Rows {

	spisok, err := r.db.Query("select get_user_segments($1)", id)
	if err != nil {
		panic(err)
	}
	return spisok

}

func (r *Repository) GetUserHistory(id int, startDate string, endDate string) *sql.Rows {

	spisok, err := r.db.Query("select * from get_user_history($1,$2,$3)", id, startDate, endDate)
	if err != nil {
		panic(err)
	}
	return spisok

}

func (r *Repository) DistributeUsers() *sql.Rows {

	spisok, err := r.db.Query("select * from get_all_users()")
	if err != nil {
		panic(err)
	}
	return spisok
}

func (r *Repository) GetSegments() *sql.Rows {

	spisok, err := r.db.Query("select * from get_all_segments()")
	if err != nil {
		panic(err)
	}
	return spisok
}
