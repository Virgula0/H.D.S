// #nosec G201 disable this rule because names of tables are statically defined as costants
package repository

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/Virgula0/progetto-dp/server/entities"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type generatedServerCerts struct {
	caCert []byte
	caKey  []byte

	serverCert []byte
	serverKey  []byte
}

type queryHandler struct {
	*sql.DB
}

type Repository struct {
	dbUser  *sql.DB
	dbCerts *sql.DB
	certs   *generatedServerCerts
}

// NewRepository Dependency Injection Pattern for injecting dbUser instance within Repository
func NewRepository(dbUser, dbCerts *infrastructure.Database) (*Repository, error) {
	return &Repository{
		dbUser:  dbUser.DB,
		dbCerts: dbCerts.DB,
		certs:   new(generatedServerCerts),
	}, nil
}

// InjectCerts Property injection on certs
func (repo *Repository) InjectCerts(caCert, caKey, serverCert, serverKey []byte) {
	repo.certs.caCert = caCert
	repo.certs.caKey = caKey
	repo.certs.serverCert = serverCert
	repo.certs.serverKey = serverKey
}

// GetServerCerts return certs
func (repo *Repository) GetServerCerts() (caCert, caKey, serverCert, serverKey []byte, err error) {
	if repo.certs.caCert == nil || repo.certs.caKey == nil ||
		repo.certs.serverCert == nil || repo.certs.serverKey == nil {
		return nil, nil, nil, nil, customErrors.ErrCertsNotInitialized
	}

	return repo.certs.caCert, repo.certs.caKey, repo.certs.serverCert, repo.certs.serverKey, nil
}

// countQueryResults function to count results
func (q *queryHandler) countQueryResults(query string, args ...any) (int, error) {

	var count int
	// Query for a value based on a single row.
	if err := q.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// queryEntities generic function for abstracting select statements in tables
func (q *queryHandler) queryEntities(query string, columnsFunc func() (any, []any), args ...any) ([]any, error) {
	var ent []any

	rows, err := q.Query(query, args...)

	if err != nil {
		log.Error(err.Error())
		return nil, customErrors.ErrInternalServerError
	}

	defer rows.Close()

	// Loop through the rows and scan into the provided entity
	for rows.Next() {
		rowEntity, rowColumns := columnsFunc()

		if err := rows.Scan(rowColumns...); err != nil {
			log.Error(err.Error())
			return nil, customErrors.ErrInternalServerError
		}
		ent = append(ent, rowEntity)
	}

	if err := rows.Err(); err != nil {
		log.Error(err.Error())
		return nil, customErrors.ErrInternalServerError
	}

	return ent, nil
}

// CreateUser creates a new record in the user and role tables
func (repo *Repository) CreateUser(userEntity *entities.User, role constants.Role) error {

	query := fmt.Sprintf("INSERT INTO %s(username, password, uuid) VALUES(?,?,?)", entities.UserTableName)

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(userEntity.Password), constants.HashCost)

	if err != nil {
		return err
	}

	_, err = repo.dbUser.Exec(query, userEntity.Username, string(passwordBytes), userEntity.UserUUID)
	if err != nil {
		return err
	}

	// Seed user role

	query = fmt.Sprintf("INSERT INTO %s(uuid,role_string) VALUES(?,?)", entities.RoleTableName)
	_, err = repo.dbUser.Exec(query, userEntity.UserUUID, role)
	if err != nil {
		return err
	}

	return nil
}

// GetUserByUsername Get a user info by username
func (repo *Repository) GetUserByUsername(username string) (*entities.User, *entities.Role, error) {

	var user entities.User
	var role entities.Role

	query := fmt.Sprintf("SELECT * FROM %s AS u NATURAL JOIN %s WHERE u.username = ? LIMIT 1", entities.UserTableName, entities.RoleTableName)

	// Execute the query expecting a single row.
	rows, err := repo.dbUser.Query(query, username)

	if err != nil {
		return nil, nil, customErrors.ErrInvalidCredentials
	}

	defer rows.Close()

	hasNext := rows.Next()

	if !hasNext {
		return nil, nil, customErrors.ErrInvalidCredentials
	}

	err = rows.Scan(&user.UserUUID, &user.Username, &user.Password, &role.RoleString)

	if err != nil {
		return nil, nil, customErrors.ErrInvalidCredentials
	}

	return &user, &role, nil
}

// GetClientsInstalled Needed in main.go for updating certs on existing clients every time the server restart
func (repo *Repository) GetClientsInstalled() (clients []*entities.Client, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s", entities.ClientTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		c := &entities.Client{}

		// Return the entity plus the columns slice
		cols := []any{
			&c.UserUUID,
			&c.ClientUUID,
			&c.Name,
			&c.LatestIP,
			&c.CreationTime,
			&c.LatestConnectionTime,
			&c.MachineID,
			&c.EnabledEncryption,
		}
		return c, cols
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(query, columnsFunc)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		client, ok := item.(*entities.Client)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Client", customErrors.ErrInvalidType)
		}
		clients = append(clients, client)
	}

	return clients, len(clients), err
}

// GetClientCertsByUserID REST-API GetClientCertsByUserID
func (repo *Repository) GetClientCertsByUserID(userUUID string) (certs []*entities.Cert, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ?", entities.CertTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		c := &entities.Cert{}

		// Return the entity plus the columns slice
		cols := []any{
			&c.CertUUID,
			&c.UserUUID,
			&c.ClientUUID,
			&c.CACert,
			&c.ClientCert,
			&c.ClientKey,
		}
		return c, cols
	}

	qq := queryHandler{repo.dbCerts}
	results, err := qq.queryEntities(query, columnsFunc, userUUID)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		cert, ok := item.(*entities.Cert)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Cert", customErrors.ErrInvalidType)
		}
		certs = append(certs, cert)
	}

	return certs, len(certs), err
}

// GetClientsInstalledByUserID REST-API GetClientsInstalledByUserID
func (repo *Repository) GetClientsInstalledByUserID(userUUID string, offset uint) (clients []*entities.Client, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.ClientTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ?", entities.ClientTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		c := &entities.Client{}

		// Return the entity plus the columns slice
		cols := []any{
			&c.UserUUID,
			&c.ClientUUID,
			&c.Name,
			&c.LatestIP,
			&c.CreationTime,
			&c.LatestConnectionTime,
			&c.MachineID,
			&c.EnabledEncryption,
		}
		return c, cols
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(query, columnsFunc, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		client, ok := item.(*entities.Client)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Client", customErrors.ErrInvalidType)
		}
		clients = append(clients, client)
	}

	count, err := qq.countQueryResults(queryCount, userUUID)

	return clients, count, err
}

// GetRaspberryPiByUserID REST-API GetRaspberryPiyUserID
func (repo *Repository) GetRaspberryPiByUserID(userUUID string, offset uint) (rsps []*entities.RaspberryPI, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.RaspberryPiTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.RaspberryPiTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		rsp := &entities.RaspberryPI{}

		// Return the entity plus the columns slice
		cols := []any{
			&rsp.UserUUID,
			&rsp.RaspberryPIUUID,
			&rsp.MachineID,
			&rsp.EncryptionKey,
		}
		return rsp, cols
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(query, columnsFunc, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		rsp, ok := item.(*entities.RaspberryPI)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.RaspberryPI", customErrors.ErrInvalidType)
		}
		rsps = append(rsps, rsp)
	}

	count, err := qq.countQueryResults(queryCount, userUUID)

	return rsps, count, err
}

// GetHandshakesByUserID REST-API - GetHandshakesByUserID returns handshakes by userID
func (repo *Repository) GetHandshakesByUserID(userUUID string, offset uint) (handshakes []*entities.Handshake, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?", entities.HandshakeTableName, constants.Limit) // TODO: remove WHERE conditions for admin roles
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? ", entities.HandshakeTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		handshake := &entities.Handshake{}

		// Return the entity plus the columns slice
		cols := []any{
			&handshake.UserUUID,
			&handshake.ClientUUID,
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
		return handshake, cols
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(query, columnsFunc, userUUID, (offset-1)*constants.Limit)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		hdk, ok := item.(*entities.Handshake)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Handshake", customErrors.ErrInvalidType)
		}
		handshakes = append(handshakes, hdk)
	}

	count, err := qq.countQueryResults(queryCount, userUUID)

	return handshakes, count, err
}

// GetHandshakesByStatus GRPC - Status of all handshake by a given filter status
func (repo *Repository) GetHandshakesByStatus(filterStatus string) (handshakes []*entities.Handshake, length int, e error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE status = ?", entities.HandshakeTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		handshake := &entities.Handshake{}

		// Return the entity plus the columns slice
		cols := []any{
			&handshake.UserUUID,
			&handshake.ClientUUID,
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
		return handshake, cols
	}
	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(query, columnsFunc, filterStatus)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		hdk, ok := item.(*entities.Handshake)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Handshake", customErrors.ErrInvalidType)
		}
		handshakes = append(handshakes, hdk)
	}

	return handshakes, len(results), err
}

// GetHandshakesByBSSIDAndSSID TCP/IP - Check if a handshake is already registered
func (repo *Repository) GetHandshakesByBSSIDAndSSID(userUUID, bssid, ssid string) (handshakes []*entities.Handshake, length int, e error) {

	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND bssid = ? AND ssid = ?", entities.HandshakeTableName)

	columnsFunc := func() (any, []any) {
		// Each time called, we make a fresh instance
		handshake := &entities.Handshake{}

		// Return the entity plus the columns slice
		cols := []any{
			&handshake.UserUUID,
			&handshake.ClientUUID,
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
		return handshake, cols
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(query, columnsFunc, userUUID, bssid, ssid)

	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		hdk, ok := item.(*entities.Handshake)
		if !ok {
			return nil, 0, fmt.Errorf("%w *entities.Handshake", customErrors.ErrInvalidType)
		}
		handshakes = append(handshakes, hdk)
	}

	return handshakes, len(results), err
}

// CreateCertForClient GRPC - Called only once when a client does not exist
func (repo *Repository) CreateCertForClient(userUUID, clientUUID string, clientCert, clientKey []byte) (string, error) {

	if repo.certs.caCert == nil || repo.certs.caKey == nil {
		return "", customErrors.ErrCertsNotInitialized
	}

	certID := uuid.New().String()
	query := fmt.Sprintf("INSERT INTO %s(uuid, uuid_user, client_uuid, ca_cert, client_cert, client_key) VALUES(?,?,?,?,?,?)", entities.CertTableName)
	_, err := repo.dbCerts.Exec(query, certID, userUUID, clientUUID, repo.certs.caCert, clientCert, clientKey)
	if err != nil {
		return "", err
	}

	return certID, nil
}

// UpdateCerts called by main.go for updating certs when server starts
func (repo *Repository) UpdateCerts(client *entities.Client, caCert, clientCert, clientKey []byte) error {
	// Update query
	updateQuery := fmt.Sprintf(
		"UPDATE %s SET ca_cert = ?, client_cert = ?, client_key = ? WHERE client_uuid = ?",
		entities.CertTableName,
	)
	_, err := repo.dbCerts.Exec(updateQuery, caCert, clientCert, clientKey, client.ClientUUID)
	if err != nil {
		return err
	}

	return nil
}

// CreateClient GRPC - CreateClient creates a new record in the client table
func (repo *Repository) CreateClient(userUUID, machineID, latestIP, name string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, name, latest_ip, creation_datetime, latest_connection, machine_id) VALUES(?,?,?,?,?,?,?)",
		entities.ClientTableName)

	formattedDateTime := time.Now().Format(constants.DateTimeExample)
	clientNewID := uuid.New().String()
	_, err := repo.dbUser.Exec(query, userUUID, clientNewID, name, latestIP, formattedDateTime, formattedDateTime, machineID)

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
		&client.EnabledEncryption,
	}

	// Execute the query expecting a single row.
	rows, err := repo.dbUser.Query(query, userUUID, machineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasNext := rows.Next()

	if !hasNext {
		return nil, customErrors.ErrNoClientFound
	}

	err = rows.Scan(columnsToBind...)

	if err != nil {
		return nil, err
	}

	return &client, nil
}

// CreateHandshake TCP/IP-REST-API - CreateHandshake creates a new record in the handshake table
func (repo *Repository) CreateHandshake(userUUID, ssid, bssid, status, handshakePcap string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, ssid, bssid, status, handshake_pcap) VALUES(?,?,?,?,?,?)",
		entities.HandshakeTableName)

	handshakeID := uuid.New().String()
	// clientID will be assigned via REST-API by the user
	// we save the userID to specify that the task can be run only from that specific user and no one else
	// in particular such userID is the same for the raspberryPI userID
	_, err := repo.dbUser.Exec(query, userUUID, handshakeID, ssid, bssid, status, handshakePcap)

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
	_, err := repo.dbUser.Exec(query, userUUID, rspNewID, machineID, encryptionKey)

	if err != nil {
		return "", err
	}

	return rspNewID, nil
}

// updateClientTaskCommon contains shared logic for updating a client task.
func (repo *Repository) updateClientTaskCommon(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake string, restMode bool) (*entities.Handshake, error) {
	selectQuery := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName)

	columnsFunc := func() (any, []any) {
		handshake := &entities.Handshake{}
		cols := []any{
			&handshake.UserUUID,
			&handshake.ClientUUID,
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
		return handshake, cols
	}

	qq := queryHandler{repo.dbUser}
	handshakes, err := qq.queryEntities(selectQuery, columnsFunc, userUUID, handshakeUUID)
	if err != nil {
		return nil, err
	}
	if len(handshakes) == 0 {
		return nil, customErrors.ErrElementNotFound
	}

	handshake, ok := handshakes[0].(*entities.Handshake)
	if !ok {
		return nil, customErrors.ErrInvalidType
	}

	// Specific REST API behavior: Check if the client is busy
	if restMode && handshake.ClientUUID != nil {
		switch handshake.Status {
		case constants.PendingStatus, constants.WorkingStatus:
			return nil, customErrors.ErrClientIsBusy
		}
	}

	// Update query
	updateQuery := fmt.Sprintf(
		"UPDATE %s SET uuid_assigned_client = ?, status = ?, hashcat_options = ?, hashcat_logs = ?, cracked_handshake = ? WHERE uuid_user = ? AND uuid = ?",
		entities.HandshakeTableName,
	)
	_, err = repo.dbUser.Exec(updateQuery, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake, userUUID, handshakeUUID)
	if err != nil {
		return nil, err
	}

	// Fetch updated handshake
	handshakes, err = qq.queryEntities(selectQuery, columnsFunc, userUUID, handshakeUUID)
	if err != nil {
		return nil, err
	}
	if len(handshakes) == 0 {
		return nil, customErrors.ErrElementNotFound
	}

	updatedHandshake, ok := handshakes[0].(*entities.Handshake)
	if !ok {
		return nil, customErrors.ErrInvalidType
	}

	return updatedHandshake, nil
}

// UpdateClientTask - GRPC version without client busy check
func (repo *Repository) UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return repo.updateClientTaskCommon(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake, false)
}

// UpdateClientTaskRest - REST version with client busy check
func (repo *Repository) UpdateClientTaskRest(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return repo.updateClientTaskCommon(userUUID, handshakeUUID, assignedClientUUID, status, haschatOptions, hashcatLogs, crackedHandshake, true)
}

// DeleteClient - REST API - Delete a client
func (repo *Repository) DeleteClient(userUUID, clientUUID string) (bool, error) {
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE uuid_user = ? AND uuid = ?", entities.ClientTableName)

	_, err := repo.dbUser.Exec(deleteQuery, userUUID, clientUUID)
	if err != nil {
		return false, fmt.Errorf("%s ERROR: %v", customErrors.ErrCannotDeleteElement, err)
	}

	return true, nil
}

// DeleteRaspberryPI - REST API - Delete a raspberry pi
func (repo *Repository) DeleteRaspberryPI(userUUID, rspUUID string) (bool, error) {
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE uuid_user = ? AND uuid = ?", entities.RaspberryPiTableName)

	_, err := repo.dbUser.Exec(deleteQuery, userUUID, rspUUID)
	if err != nil {
		return false, fmt.Errorf("%s ERROR: %v", customErrors.ErrCannotDeleteElement, err)
	}

	return true, nil
}

// DeleteHandshake - REST API - Delete an handshake
func (repo *Repository) DeleteHandshake(userUUID, handshakeUUID string) (bool, error) {
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName)

	_, err := repo.dbUser.Exec(deleteQuery, userUUID, handshakeUUID)
	if err != nil {
		return false, fmt.Errorf("%s ERROR: %v", customErrors.ErrCannotDeleteElement, err)
	}

	return true, nil
}
