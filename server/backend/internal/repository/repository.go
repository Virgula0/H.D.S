// #nosec G201 disable this rule because names of tables are statically defined as costants
package repository

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	"github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/Virgula0/progetto-dp/server/entities"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sql.DB
}

// NewRepository Dependency Injection Pattern for injecting db instance within Repository
func NewRepository(db *infrastructure.Database) (*Repository, error) {
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

// GetUserByUsername Get an user info by username
func (repo *Repository) GetUserByUsername(username string) (*entities.User, *entities.Role, error) {

	var user entities.User
	var role entities.Role

	query := fmt.Sprintf("SELECT * FROM %s AS u NATURAL JOIN %s WHERE u.username = ? LIMIT 1", entities.UserTableName, entities.RoleTableName)

	// Execute the query expecting a single row.
	rows, err := repo.db.Query(query, username)

	if err != nil {
		return nil, nil, errors.ErrInvalidCredentials
	}

	defer rows.Close()

	hasNext := rows.Next()

	if !hasNext {
		return nil, nil, errors.ErrInvalidCredentials
	}

	err = rows.Scan(&user.UserUUID, &user.Username, &user.Password, &role.RoleString)

	if err != nil {
		return nil, nil, errors.ErrInvalidCredentials
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

// queryEntities generic function for abstracting select statements in tables
func (repo *Repository) queryEntities(query string, columns []any, entity any, args ...any) ([]any, error) {
	var ent []any

	rows, err := repo.db.Query(query, args...)

	if err != nil {
		log.Error(err.Error())
		return nil, errors.ErrInternalServerError
	}
	defer rows.Close()

	// Loop through the rows and scan into the provided entity
	for rows.Next() {
		if err := rows.Scan(columns...); err != nil {
			log.Error(err.Error())
			return nil, errors.ErrInternalServerError
		}
		ent = append(ent, entity)
	}

	if err := rows.Err(); err != nil {
		log.Error(err.Error())
		return nil, errors.ErrInternalServerError
	}

	return ent, nil
}

// GetClientsInstalledByUserID REST-API GetClientsInstalledByUserID
func (repo *Repository) GetClientsInstalledByUserID(userUUID string, offset uint) (clients []*entities.Client, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.ClientTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.ClientTableName)

	var client entities.Client

	columnsToBind := []any{
		&client.UserUUID,
		&client.ClientUUID,
		&client.Name,
		&client.LatestIP,
		&client.CreationTime,
		&client.LatestConnectionTime,
		&client.MachineID,
	}

	results, err := repo.queryEntities(query, columnsToBind, &client, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		client, ok := item.(*entities.Client)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Client", errors.ErrInvalidType)
		}
		clients = append(clients, client)
	}

	count, err := repo.countQueryResults(queryCount, userUUID)

	return clients, count, err
}

// GetRaspberryPiByUserID REST-API GetRaspberryPiyUserID
func (repo *Repository) GetRaspberryPiByUserID(userUUID string, offset uint) (rsps []*entities.RaspberryPI, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.RaspberryPiTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.RaspberryPiTableName)

	var rsp entities.RaspberryPI

	columnsToBind := []any{
		&rsp.UserUUID,
		&rsp.RaspberryPIUUID,
		&rsp.MachineID,
		&rsp.EncryptionKey,
	}

	results, err := repo.queryEntities(query, columnsToBind, &rsp, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		rsp, ok := item.(*entities.RaspberryPI)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.RaspberryPI", errors.ErrInvalidType)
		}
		rsps = append(rsps, rsp)
	}

	count, err := repo.countQueryResults(queryCount, userUUID)

	return rsps, count, err
}

// GetHandshakesByUserID REST-API - GetHandshakesByUserID returns handshakes by userID
func (repo *Repository) GetHandshakesByUserID(userUUID string, offset uint) (handshakes []*entities.Handshake, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.HandshakeTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.HandshakeTableName)

	var handshake entities.Handshake

	columnsToBind := []any{
		&handshake.UserUUID,
		&handshake.ClientUUID,
		&handshake.RaspberryPIUUID,
		&handshake.UUID,
		&handshake.SSID,
		&handshake.BSSID,
		&handshake.UploadedDate,
		&handshake.Status,
		&handshake.CrackedDate,
		&handshake.HashcatOptions,
		&handshake.HashcatLogs,
		&handshake.CrackedHandshake,
		&handshake.HandshakePCAP,
	}

	results, err := repo.queryEntities(query, columnsToBind, &handshake, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		hdk, ok := item.(*entities.Handshake)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Handshake", errors.ErrInvalidType)
		}
		handshakes = append(handshakes, hdk)
	}

	count, err := repo.countQueryResults(queryCount, userUUID)

	return handshakes, count, err
}

// GetHandshakesByStatus GRPC - Status of all handshake by a given filter status
func (repo *Repository) GetHandshakesByStatus(filterStatus string) (handshakes []*entities.Handshake, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE status = ?", entities.HandshakeTableName)
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE status = ? ", entities.HandshakeTableName)

	var handshake entities.Handshake

	columnsToBind := []any{
		&handshake.UserUUID,
		&handshake.ClientUUID,
		&handshake.RaspberryPIUUID,
		&handshake.UUID,
		&handshake.SSID,
		&handshake.BSSID,
		&handshake.UploadedDate,
		&handshake.Status,
		&handshake.CrackedDate,
		&handshake.HashcatOptions,
		&handshake.HashcatLogs,
		&handshake.CrackedHandshake,
		&handshake.HandshakePCAP,
	}

	results, err := repo.queryEntities(query, columnsToBind, &handshake, filterStatus)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		hdk, ok := item.(*entities.Handshake)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Handshake", errors.ErrInvalidType)
		}
		handshakes = append(handshakes, hdk)
	}

	count, err := repo.countQueryResults(queryCount, filterStatus)

	return handshakes, count, err
}

// CreateClient GRPC - CreateClient creates a new record in the client table
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

// GetClientInfo GRPC - GetClientInfo get client info by userID and machineID
func (repo *Repository) GetClientInfo(userUUID, machineID string) (*entities.Client, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? and machine_id = ?", entities.ClientTableName)

	var client entities.Client

	columnsToBind := []any{
		&client.UserUUID,
		&client.ClientUUID,
		&client.Name,
		&client.LatestIP,
		&client.CreationTime,
		&client.LatestConnectionTime,
		&client.MachineID,
	}

	// Execute the query expecting a single row.
	rows, err := repo.db.Query(query, userUUID, machineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasNext := rows.Next()

	if !hasNext {
		return nil, errors.ErrNoClientFound
	}

	err = rows.Scan(columnsToBind...)

	if err != nil {
		return nil, err
	}

	return &client, nil
}

// CreateHandshake TCP/IP - CreateHandshake creates a new record in the handshake table
func (repo *Repository) CreateHandshake(userUUID, raspberryPIUUID, ssid, bssid, status, handshakePcap string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid_assigned_raspberry_pi, uuid, ssid, bssid, status, handshake_pcap) VALUES(?,?,?,?,?,?,?)",
		entities.HandshakeTableName)

	handshakeID := uuid.New().String()
	// clientID will be assigned via REST-API by the user
	// we save the userID to specify that the task can be run only from that specific user and no one else
	// in particular such userID is the same for the raspberryPI userID
	_, err := repo.db.Exec(query, userUUID, raspberryPIUUID, handshakeID, ssid, bssid, status, handshakePcap)

	if err != nil {
		return "", err
	}

	return handshakeID, nil
}

// CreateRaspberryPI TCP/IP Server - CreateRaspberryPI creates a new raspberry-pi device entry
func (repo *Repository) CreateRaspberryPI(userUUID, machineID, encryptionKey string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, machine_id, encryption_key) VALUES(?,?,?,?)",
		entities.RaspberryPiTableName)

	rspNewID := uuid.New().String()
	_, err := repo.db.Exec(query, userUUID, rspNewID, machineID, encryptionKey)

	if err != nil {
		return "", err
	}

	return rspNewID, nil
}

// UpdateClientTask REST-API/GRPC - UpdateClientTask updates a new client task
func (repo *Repository) UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	selectQuery := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName)

	// Define a variable for the handshake
	var handshake entities.Handshake

	// Bind columns for querying
	columnsToBind := []any{
		&handshake.UserUUID,
		&handshake.ClientUUID,
		&handshake.RaspberryPIUUID,
		&handshake.UUID,
		&handshake.SSID,
		&handshake.BSSID,
		&handshake.UploadedDate,
		&handshake.Status,
		&handshake.CrackedDate,
		&handshake.HashcatOptions,
		&handshake.HashcatLogs,
		&handshake.CrackedHandshake,
		&handshake.HandshakePCAP,
	}

	handshakes, err := repo.queryEntities(selectQuery, columnsToBind, &handshake, userUUID, handshakeUUID)
	if err != nil {
		return nil, err
	}

	if len(handshakes) == 0 {
		return nil, errors.ErrElementNotFound
	}

	// Ensure that the result is of type Handshake
	if _, ok := handshakes[0].(*entities.Handshake); !ok {
		return nil, errors.ErrInvalidType
	}

	updateQuery := fmt.Sprintf("UPDATE %s SET uuid_assigned_client = ?, status = ?, hashcat_options = ?, hashcat_logs = ?, cracked_handshake = ? WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName)
	_, err = repo.db.Exec(updateQuery, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake, userUUID, handshakeUUID)
	if err != nil {
		return nil, err
	}

	// Get updated data
	handshakes, err = repo.queryEntities(selectQuery, columnsToBind, &handshake, userUUID, handshakeUUID)
	if err != nil {
		return nil, err
	}

	if len(handshakes) == 0 {
		return nil, errors.ErrElementNotFound
	}

	converted, ok := handshakes[0].(*entities.Handshake)
	if !ok {
		return nil, errors.ErrInvalidType
	}

	return converted, nil
}
