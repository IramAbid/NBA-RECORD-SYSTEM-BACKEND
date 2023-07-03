package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all of the types we want to be available to our application.
func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		Student: Student{},
		Token:   Token{},
	}
}

// Models is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, provided that the model is also added in the New function.
type Models struct {
	Student Student
	Token   Token
}

// Student is the stucture which holds one Student from the database. Note
// that it embeds a token type.
type Student struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"first_name,omitempty"`
	LastName          string    `json:"last_name,omitempty"`
	EnrollmentNo      string    `json:"enrollment_no,omitempty"`
	FacultyNo         string    `json:"faculty_no,omitempty"`
	Course            string    `json:"course,omitempty"`
	Semester          int       `json:"semester,omitempty"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	RegisteredCourses []string  `json:"registered_courses,omitempty"`
	Feedbacks         []string  `json:"feedbacks,omitempty"`
	FilledFeedbacks   []string  `json:"filled_feedbacks,omitempty"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// for token put it above

// GetAll returns a slice of all Students, sorted by last name
func (u *Student) GetAll() ([]*Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, enrollment_no, faculty_no, course, semester from students order by last_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Students []*Student

	for rows.Next() {
		var Student Student
		err := rows.Scan(
			&Student.ID,
			&Student.Email,
			&Student.FirstName,
			&Student.LastName,
			&Student.EnrollmentNo,
			&Student.FacultyNo,
			&Student.Course,
			&Student.Semester,
		)
		if err != nil {
			return nil, err
		}

		Students = append(Students, &Student)
	}

	return Students, nil
}

// GetByEmail returns one Student by email
func (u *Student) GetByEmail(email string) (*Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, enrollment_no, password from Students where email = $1`

	var Student Student
	row := db.QueryRowContext(ctx, query, email)
	err := row.Scan(
		&Student.ID,
		&Student.Email,
		&Student.FirstName,
		&Student.LastName,
		&Student.EnrollmentNo,
		&Student.Password,
	)

	if err != nil {
		return nil, err
	}

	return &Student, nil
}

// GetOne returns one Student by id
func (u *Student) GetOne(id int) (*Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from Students where id = $1`

	var Student Student
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&Student.ID,
		&Student.Email,
		&Student.FirstName,
		&Student.LastName,
		&Student.Password,
		&Student.CreatedAt,
		&Student.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &Student, nil
}

// Update updates one Student in the database, using the information
// stored in the receiver u
func (u *Student) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `update Students set
		email = $1,
		first_name = $2,
		last_name = $3,
		updated_at = $4
		where id = $5
	`

	_, err := db.ExecContext(ctx, stmt,
		u.Email,
		u.FirstName,
		u.LastName,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes one Student from the datbase, by ID
func (u *Student) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from Students where id = $1`

	_, err := db.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new Student into the datbase, and returns the ID of the newly inserted row
func (u *Student) Insert(Student Student) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Student.Password), 12)
	if err != nil {
		return 0, err
	}

	var newID int
	stmt := `insert into Students (email, first_name, last_name, password, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6) returning id`

	err = db.QueryRowContext(ctx, stmt,
		Student.Email,
		Student.FirstName,
		Student.LastName,
		hashedPassword,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// ResetPassword is the method we will use to change a Student's password.
func (u *Student) ResetPassword(password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `update Students set password = $1 where id = $2`
	_, err = db.ExecContext(ctx, stmt, hashedPassword, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// PasswordMatches uses Go's bcrypt package to compare a Student supplied password
// with the hash we have stored for a given Student in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (u *Student) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (u *Student) GetCourses(id int) (*Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `select registered_courses, feedbacks, filled_feedbacks from students where id=$1`

	row := db.QueryRowContext(ctx, stmt, id)

	var Student Student

	err := row.Scan(
		pq.Array(&Student.RegisteredCourses),
		pq.Array(&Student.Feedbacks),
		pq.Array(&Student.FilledFeedbacks),
	)
	if err!=nil{
		return nil, err;
	}

	return &Student, nil

}

// Token is the data structure for any token in the database. Note that
// we do not send the TokenHash (a slice of bytes) in any exported JSON.
type Token struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	TokenHash []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Expiry    time.Time `json:"expiry"`
}

// GetByToken takes a plain text token string, and looks up the full token from
// the database. It returns a pointer to the Token model.
func (t *Token) GetByToken(plainText string) (*Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, student_id, email, token, token_hash, created_at, updated_at, expiry
			from tokens where token = $1`

	var token Token

	row := db.QueryRowContext(ctx, query, plainText)
	err := row.Scan(
		&token.ID,
		&token.StudentID,
		&token.Email,
		&token.Token,
		&token.TokenHash,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.Expiry,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

// GetStudentForToken takes a token parameter, and uses the StudentID field from that parameter
// to look a Student up by id. It returns a pointer to the Student model.
func (t *Token) GetStudentForToken(token Token) (*Student, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from Students where id = $1`

	var Student Student
	row := db.QueryRowContext(ctx, query, token.StudentID)

	err := row.Scan(
		&Student.ID,
		&Student.Email,
		&Student.FirstName,
		&Student.LastName,
		&Student.Password,
		&Student.CreatedAt,
		&Student.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &Student, nil
}

// GenerateToken generates a secure token of exactly 26 characters in length and returns it
func (t *Token) GenerateToken(user Student, ttl time.Duration) (*Token, error) {
	token := &Token{
		StudentID: user.ID,
		Email:     user.Email,
		Expiry:    time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Token = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Token))
	token.TokenHash = hash[:]

	return token, nil
}

// AuthenticateToken takes the full http request, extracts the authorization header,
// takes the plain text token from that header and looks up the associated token entry
// in the database, and then finds the Student associated with that token. If the token
// is valid and a Student is found, the Student is returned; otherwise, it returns an error.
func (t *Token) AuthenticateToken(r *http.Request) (*Student, error) {
	// get the authorization header
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header received")
	}

	// get the plain text token from the header
	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no valid authorization header received")
	}

	token := headerParts[1]

	// make sure the token is of the correct length
	if len(token) != 26 {
		return nil, errors.New("token wrong size")
	}

	// get the token from the database, using the plain text token to find it
	tkn, err := t.GetByToken(token)
	if err != nil {
		return nil, errors.New("no matching token found")
	}

	// make sure the token has not expired
	if tkn.Expiry.Before(time.Now()) {
		return nil, errors.New("expired token")
	}

	// get the Student associated with the token
	Student, err := t.GetStudentForToken(*tkn)
	if err != nil {
		return nil, errors.New("no matching Student found")
	}

	return Student, nil
}

// Insert inserts a token into the database
func (t *Token) Insert(token Token, u Student) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// delete any existing tokens
	stmt := `delete from tokens where Student_id = $1`
	_, err := db.ExecContext(ctx, stmt, token.StudentID)
	if err != nil {
		return err
	}

	// we assign the email value, just to be safe, in case it was
	// not done in the handler that calls this function
	token.Email = u.Email

	// insert the new token
	stmt = `insert into tokens (student_id, email, token, token_hash, created_at, updated_at, expiry)
		values ($1, $2, $3, $4, $5, $6, $7)`

	_, err = db.ExecContext(ctx, stmt,
		u.ID,
		u.Email,
		token.Token,
		token.TokenHash,
		time.Now(),
		time.Now(),
		token.Expiry,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteByToken deletes a token, by plain text token
func (t *Token) DeleteByToken(plainText string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from tokens where token = $1`

	_, err := db.ExecContext(ctx, stmt, plainText)
	if err != nil {
		return err
	}

	return nil
}

// ValidToken makes certain that a given token is valid; in order to be valid,
// the token must exist in the database, the associated Student must exist in the database,
// and the token must not have expired.
func (t *Token) ValidToken(plainText string) (bool, error) {
	token, err := t.GetByToken(plainText)
	if err != nil {
		return false, errors.New("no matching token found")
	}

	_, err = t.GetStudentForToken(*token)
	if err != nil {
		return false, errors.New("no matching Student found")
	}

	if token.Expiry.Before(time.Now()) {
		return false, errors.New("expired token")
	}

	return true, nil
}
