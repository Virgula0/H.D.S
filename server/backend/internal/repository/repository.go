package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/entities"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sql.DB
}

// Dependency Injection Pattern for injecting db instance within Repository
func NewRepository(db *infrastructure.Database, reset bool) (*Repository, error) {
	return &Repository{
		db.DB,
	}, nil
}

// CreateUser creates a new record in the user and role tables
func (repo *Repository) CreateUser(userEntity *entities.User, role constants.Role) error {

	query := fmt.Sprintf("INSERT INTO %s(username, password, uuid) VALUES(?,?,?)", entities.UserTableName)

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(userEntity.Password), constants.HashCost)

	if err != nil {
		return err
	}

	_, err = repo.db.Exec(query, userEntity.Username, string(passwordBytes), userEntity.UserUUID)
	if err != nil {
		return err
	}

	// Seed user role

	query = fmt.Sprintf("INSERT INTO %s(uuid,role_string) VALUES(?,?)", entities.RoleTableName)
	_, err = repo.db.Exec(query, userEntity.UserUUID, role)
	if err != nil {
		return err
	}

	return nil
}

// CreateUser creates a new record in the user table
func (repo *Repository) GetUserByUsername(username string) (*entities.User, *entities.Role, error) {

	var user entities.User
	var role entities.Role

	query := fmt.Sprintf("SELECT * FROM %s AS u NATURAL JOIN %s WHERE u.username = ? LIMIT 1", entities.UserTableName, entities.RoleTableName)

	// Execute the query expecting a single row.
	rows, err := repo.db.Query(query, username)

	if err != nil {
		return nil, nil, errors.ErrInvalidCreds
	}

	defer rows.Close()

	hasNext := rows.Next()

	if !hasNext {
		return nil, nil, errors.ErrInvalidCreds
	}

	err = rows.Scan(&user.UserUUID, &user.Username, &user.Password, &role.RoleString)

	if err != nil {
		return nil, nil, errors.ErrInvalidCreds
	}

	return &user, &role, nil
}

// countQueryResults function to count results
func (repo *Repository) countQueryResults(query string, args ...any) (int, error) {

	var count int
	// Query for a value based on a single row.
	if err := repo.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil

}

func (repo *Repository) queryClients(query string, columns []any, client *entities.Client, args ...any) ([]*entities.Client, error) {
	var clients []*entities.Client

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		//var client entities.Client
		if err = rows.Scan(columns...); err != nil {
			return nil, errors.ErrInternalServerError
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.ErrInternalServerError
	}

	return clients, nil
}

// REST-API GetClientsInstalledByUser
func (repo *Repository) GetClientsInstalledByUser(userUUID string, offset uint) ([]*entities.Client, int, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.ClientTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.ClientTableName)

	var client entities.Client

	columns := []any{
		&client.UserUUID,
		&client.ClientUUID,
		&client.Name,
		&client.LatestIP,
		&client.CreationTime,
		&client.LatestConnectionTime,
		&client.MachineID,
	}

	clients, err := repo.queryClients(query, columns, &client, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	count, err := repo.countQueryResults(queryCount, userUUID)

	return clients, count, err
}

// GRPC - CreatePost creates a new record in the post table
func (repo *Repository) CreateClient(userUUID, machineID, latestIP, name string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, name, latest_ip, creation_datetime, latest_connection, machine_id) VALUES(?,?,?,?,?,?,?)",
		entities.ClientTableName)

	formattedDateTime := time.Now().Format(constants.DateTimeExample)
	clientNewID := uuid.New().String()
	_, err := repo.db.Exec(query, userUUID, clientNewID, name, latestIP, formattedDateTime, formattedDateTime, machineID)

	if err != nil {
		return "", err
	}

	return clientNewID, nil
}

// TODO: remove these commented
/*
// countQueryResults function to count results
func (repo *Repository) countQueryResults(query string, args ...any) (int, error) {

	var count int
	// Query for a value based on a single row.
	if err := repo.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil

}

// CreatePost creates a new record in the post table
func (repo *Repository) CreatePost(userUUID, title, content string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid_post, title, content, creation_datetime, latest_revision_datetime) VALUES(?,?,?,?,?,?)",
		entities.PostTableName)

	formattedDateTime := time.Now().Format(constants.DateTimeExample)
	postNewID := uuid.New().String()
	_, err := repo.db.Exec(query, userUUID, postNewID, title, utils.StringToBase64String(content), formattedDateTime, formattedDateTime)

	if err != nil {
		return "", err
	}

	return postNewID, nil
}

// Common private method for executing queries and scanning rows into Post entities
func (repo *Repository) queryPosts(query string, args ...any) ([]*entities.Post, error) {
	var posts []*entities.Post

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, errors.ErrInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var post entities.Post
		if err = rows.Scan(&post.UUIDUser, &post.UUIDPost, &post.Title, &post.Content, &post.CreationDateTime, &post.LatestRevisionDateTime); err != nil {
			return nil, errors.ErrInternalServerError
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.ErrInternalServerError
	}

	return posts, nil
}

// GetUserPosts function
func (repo *Repository) GetUserPosts(userUUID string, offset uint) ([]*entities.Post, int, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.PostTableName, constants.Limit)
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.PostTableName)

	posts, err := repo.queryPosts(query, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	count, err := repo.countQueryResults(queryCount, userUUID)

	return posts, count, err
}

// GetAllPosts created. Intended for ADMINs only
func (repo *Repository) GetAllPosts(offset uint) ([]*entities.Post, int, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY RAND() LIMIT %v OFFSET ?", entities.PostTableName, constants.Limit)
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s", entities.PostTableName)

	posts, err := repo.queryPosts(query, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	count, err := repo.countQueryResults(queryCount)

	return posts, count, err
}

// GetPostByID get a post information by passing an uuid
func (repo *Repository) GetPostByID(postUUID string) (*entities.Post, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_post = ?", entities.PostTableName)
	posts, err := repo.queryPosts(query, postUUID)

	if len(posts) > 0 {
		return posts[0], err // it should be unique
	}

	return nil, errors.ErrElementNotFound
}

func (repo *Repository) EditPostByUser(userUUID, postUUID, title, content string) (*entities.Post, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_post = ? AND uuid_user = ?", entities.PostTableName)
	posts, err := repo.queryPosts(query, postUUID, userUUID)

	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, errors.ErrElementNotFound
	}

	post := posts[0] // it should be unique anyway

	query = fmt.Sprintf("UPDATE %s SET title = ?, content = ? WHERE uuid_post = ? AND uuid_user = ?", entities.PostTableName)
	_, err = repo.db.Exec(query, title, content, post.UUIDPost, post.UUIDUser)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *Repository) UpdateLatestRevision(postUUID string) (*entities.Post, error) {

	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_post = ?", entities.PostTableName)
	posts, err := repo.queryPosts(query, postUUID)

	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, errors.ErrElementNotFound
	}

	post := posts[0] // it should be unique anyway

	formattedDateTime := time.Now().Format(constants.DateTimeExample)

	query = fmt.Sprintf("UPDATE %s SET latest_revision_datetime = ? WHERE uuid_post = ?", entities.PostTableName)
	_, err = repo.db.Exec(query, formattedDateTime, post.UUIDPost)
	if err != nil {
		return nil, err
	}

	return post, nil
}
*/
