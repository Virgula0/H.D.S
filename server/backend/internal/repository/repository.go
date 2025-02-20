// #nosec G201 for SQL false positives
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Virgula0/progetto-dp/server/backend/internal/constants"
	customErrors "github.com/Virgula0/progetto-dp/server/backend/internal/errors"
	"github.com/Virgula0/progetto-dp/server/backend/internal/infrastructure"
	"github.com/Virgula0/progetto-dp/server/entities"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Repository handles database operations using two separate connections
type Repository struct {
	dbUser  *sql.DB
	dbCerts *sql.DB
	certs   *generatedServerCerts
}

type generatedServerCerts struct {
	caCert     []byte
	caKey      []byte
	serverCert []byte
	serverKey  []byte
}

type queryHandler struct {
	*sql.DB
}

// NewRepository creates a new Repository instance with injected database connections
func NewRepository(dbUser, dbCerts *infrastructure.Database) (*Repository, error) {
	return &Repository{
		dbUser:  dbUser.DB,
		dbCerts: dbCerts.DB,
		certs:   new(generatedServerCerts),
	}, nil
}

// InjectCerts injects TLS certificates into the repository
func (repo *Repository) InjectCerts(caCert, caKey, serverCert, serverKey []byte) {
	repo.certs.caCert = caCert
	repo.certs.caKey = caKey
	repo.certs.serverCert = serverCert
	repo.certs.serverKey = serverKey
}

// GetServerCerts returns stored TLS certificates
func (repo *Repository) GetServerCerts() (caCert, caKey, serverCert, serverKey []byte, err error) {
	if repo.certs.caCert == nil || repo.certs.caKey == nil ||
		repo.certs.serverCert == nil || repo.certs.serverKey == nil {
		return nil, nil, nil, nil, customErrors.ErrCertsNotInitialized
	}
	return repo.certs.caCert, repo.certs.caKey, repo.certs.serverCert, repo.certs.serverKey, nil
}

// countQueryResults executes a count query and returns the result
func (q *queryHandler) countQueryResults(query string, args ...any) (int, error) {
	var count int
	if err := q.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}
	return count, nil
}

// queryEntities executes a query and returns scanned entities using generic scanning
func (q *queryHandler) queryEntities(query string, builder func() (any, []any), args ...any) ([]any, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		log.Error("Query execution error: ", err.Error())
		return nil, customErrors.ErrInternalServerError
	}

	// Use type-asserted scanRows wrapper to handle interface conversion
	results, err := scanRowsInterfaceWrapper(rows, builder)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// scanRowsInterfaceWrapper handles interface type conversion for generic scanRows
func scanRowsInterfaceWrapper(rows *sql.Rows, builder func() (any, []any)) ([]any, error) {
	defer rows.Close()
	var results []any

	for rows.Next() {
		entity, dest := builder()
		if err := rows.Scan(dest...); err != nil {
			log.Error("Row scan error: ", err.Error())
			return nil, customErrors.ErrInternalServerError
		}
		results = append(results, entity)
	}

	if err := rows.Err(); err != nil {
		log.Error("Rows iteration error: ", err.Error())
		return nil, customErrors.ErrInternalServerError
	}
	return results, nil
}

// CreateUser creates a new record in the user and role tables
func (repo *Repository) CreateUser(userEntity *entities.User, role constants.Role) error {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(userEntity.Password), constants.HashCost)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// User insert
	userQuery := fmt.Sprintf("INSERT INTO %s(username, password, uuid) VALUES(?,?,?)", entities.UserTableName)
	if _, err := repo.dbUser.Exec(userQuery, userEntity.Username, string(passwordBytes), userEntity.UserUUID); err != nil {
		return fmt.Errorf("user insert failed: %w", err)
	}

	// Role insert (maintain original non-transactional approach)
	roleQuery := fmt.Sprintf("INSERT INTO %s(uuid,role_string) VALUES(?,?)", entities.RoleTableName)
	if _, err := repo.dbUser.Exec(roleQuery, userEntity.UserUUID, role); err != nil {
		return fmt.Errorf("role insert failed: %w", err)
	}

	return nil
}

// GetUserByUsername retrieves user and role information by username
func (repo *Repository) GetUserByUsername(username string) (*entities.User, *entities.Role, error) {
	var user entities.User
	var role entities.Role

	query := fmt.Sprintf("SELECT * FROM %s AS u NATURAL JOIN %s WHERE u.username = ? LIMIT 1",
		entities.UserTableName, entities.RoleTableName)

	row := repo.dbUser.QueryRow(query, username)
	err := row.Scan(&user.UserUUID, &user.Username, &user.Password, &role.RoleString)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, customErrors.ErrInvalidCredentials
		}
		log.Error("User query error: ", err.Error())
		return nil, nil, customErrors.ErrInternalServerError
	}

	return &user, &role, nil
}

// GetUserByUserID REST/API retrives the user by userID
func (repo *Repository) GetUserByUserID(userUUID string) (*entities.User, error) {
	var user entities.User

	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid = ?", entities.UserTableName)

	row := repo.dbUser.QueryRow(query, userUUID)
	err := row.Scan(&user.UserUUID, &user.Username, &user.Password)

	return &user, err
}

// GetClientsInstalled returns all installed clients
func (repo *Repository) GetClientsInstalled() (clients []*entities.Client, length int, e error) {
	clientBuilder := func() (any, []any) {
		c := &entities.Client{}
		return c, []any{
			&c.UserUUID,
			&c.ClientUUID,
			&c.Name,
			&c.LatestIP,
			&c.CreationTime,
			&c.LatestConnectionTime,
			&c.MachineID,
			&c.EnabledEncryption,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s", entities.ClientTableName),
		clientBuilder,
	)
	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		clients = append(clients, item.(*entities.Client))
	}
	return clients, len(clients), nil
}

// GetWordlistByClientUUID returns wordlist for a given client uuid
func (repo *Repository) GetWordlistByClientUUID(userUUID, clientUUID string) (list []*entities.Wordlist, length int, err error) {
	wordlistBuilder := func() (any, []any) {
		c := &entities.Wordlist{}
		return c, []any{
			&c.UUID,
			&c.UserUUID,
			&c.ClientUUID,
			&c.WordlistName,
			&c.WordlistHash,
			&c.WordlistSize,
			&c.WordlistFileContent,
			&c.WordlistLocationPath,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND client_uuid = ?", entities.WordlistTableName),
		wordlistBuilder,
		userUUID,
		clientUUID,
	)
	if err != nil {
		return nil, -1, err
	}

	var listResult []*entities.Wordlist

	// Convert interface slice to concrete type
	for _, item := range results {
		listResult = append(listResult, item.(*entities.Wordlist))
	}

	count, err := qq.countQueryResults(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? AND client_uuid = ? ", entities.WordlistTableName),
		userUUID,
		clientUUID,
	)
	return listResult, count, err
}

// GetWordlistByClientAndWordlistUUID returns wordlist for a given client uuid
func (repo *Repository) GetWordlistByClientAndWordlistUUID(userUUID, clientUUID, wordlistUUID string) (*entities.Wordlist, error) {
	var ww entities.Wordlist
	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND client_uuid = ? AND uuid = ?", entities.WordlistTableName)

	row := repo.dbUser.QueryRow(query, userUUID, clientUUID, wordlistUUID)
	if err := row.Scan(&ww.UUID, &ww.UserUUID, &ww.ClientUUID, &ww.WordlistName, &ww.WordlistHash, &ww.WordlistSize, &ww.WordlistFileContent, &ww.WordlistLocationPath); err != nil {
		return nil, err
	}

	return &ww, nil
}

// CreateWordlist gRPC
func (repo *Repository) CreateWordlist(wordlistEntity *entities.Wordlist) error {
	qq := queryHandler{repo.dbUser}
	// check if exists first
	count, err := qq.countQueryResults(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ? AND wordlist_hash = ? ", entities.WordlistTableName),
		wordlistEntity.UserUUID,
		wordlistEntity.WordlistHash,
	)

	if err != nil {
		return err
	}

	if count > 0 {
		return customErrors.ErrWordlistAlreadyPresent
	}

	// Wordlist insert
	userQuery := fmt.Sprintf("INSERT INTO %s(uuid, uuid_user, client_uuid, wordlist_name, wordlist_hash, wordlist_size, file_content ) VALUES(?,?,?,?,?,?,?)", entities.WordlistTableName)
	if _, err := repo.dbUser.Exec(userQuery,
		uuid.New().String(),
		wordlistEntity.UserUUID,
		wordlistEntity.ClientUUID,
		wordlistEntity.WordlistName,
		wordlistEntity.WordlistHash,
		wordlistEntity.WordlistSize,
		wordlistEntity.WordlistFileContent,
	); err != nil {
		return fmt.Errorf("wordlist insert failed: %w", err)
	}
	return nil
}

// GetClientCertsByUserID returns client certificates for a user
func (repo *Repository) GetClientCertsByUserID(userUUID string) (certs []*entities.Cert, length int, e error) {
	certBuilder := func() (any, []any) {
		c := &entities.Cert{}
		return c, []any{
			&c.CertUUID,
			&c.UserUUID,
			&c.ClientUUID,
			&c.CACert,
			&c.ClientCert,
			&c.ClientKey,
		}
	}

	qq := queryHandler{repo.dbCerts}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ?", entities.CertTableName),
		certBuilder,
		userUUID,
	)
	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		certs = append(certs, item.(*entities.Cert))
	}
	return certs, len(certs), nil
}

// GetClientsInstalledByUserID returns paginated clients for a user
func (repo *Repository) GetClientsInstalledByUserID(userUUID string, offset uint) (clients []*entities.Client, length int, e error) {
	clientBuilder := func() (any, []any) {
		c := &entities.Client{}
		return c, []any{
			&c.UserUUID,
			&c.ClientUUID,
			&c.Name,
			&c.LatestIP,
			&c.CreationTime,
			&c.LatestConnectionTime,
			&c.MachineID,
			&c.EnabledEncryption,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?",
			entities.ClientTableName, constants.Limit),
		clientBuilder,
		userUUID, (offset-1)*constants.Limit,
	)
	if err != nil {
		return nil, -1, err
	}

	// Convert interface slice to concrete type
	for _, item := range results {
		clients = append(clients, item.(*entities.Client))
	}

	count, err := qq.countQueryResults(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ?", entities.ClientTableName),
		userUUID,
	)
	return clients, count, err
}

// UpdateEncryptionClientStatus updates client encryption status
func (repo *Repository) UpdateEncryptionClientStatus(clientUUID, userUUID string, status bool) error {
	updateQuery := fmt.Sprintf(
		"UPDATE %s SET enabled_encryption = ? WHERE uuid_user = ? AND uuid = ?",
		entities.ClientTableName,
	)
	_, err := repo.dbUser.Exec(updateQuery, status, userUUID, clientUUID)
	return err
}

// CreateClient creates a new client record
func (repo *Repository) CreateClient(userUUID, machineID, latestIP, name string) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, name, latest_ip, creation_datetime, latest_connection, machine_id) VALUES(?,?,?,?,?,?,?)",
		entities.ClientTableName)

	formattedDateTime := time.Now().Format(constants.DateTimeExample)
	clientNewID := uuid.New().String()
	_, err := repo.dbUser.Exec(query, userUUID, clientNewID, name, latestIP, formattedDateTime, formattedDateTime, machineID)
	return clientNewID, err
}

// UpdateClientTaskCommon contains shared logic for updating client tasks
func (repo *Repository) updateClientTaskCommon(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake string, restMode bool) (*entities.Handshake, error) {
	handshakeBuilder := func() (any, []any) {
		h := &entities.Handshake{}
		return h, []any{
			&h.UserUUID,
			&h.ClientUUID,
			&h.UUID,
			&h.SSID,
			&h.BSSID,
			&h.UploadedDate,
			&h.Status,
			&h.CrackedDate,
			&h.HashcatOptions,
			&h.HashcatLogs,
			&h.CrackedHandshake,
			&h.HandshakePCAP,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName),
		handshakeBuilder,
		userUUID, handshakeUUID,
	)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, customErrors.ErrElementNotFound
	}

	handshake, ok := results[0].(*entities.Handshake)
	if !ok {
		return nil, customErrors.ErrInvalidType
	}

	// Validate client status for REST mode
	if restMode && handshake.ClientUUID != nil {
		switch handshake.Status {
		case constants.PendingStatus, constants.WorkingStatus:
			return nil, customErrors.ErrClientIsBusy
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE %s SET uuid_assigned_client = ?, status = ?, hashcat_options = ?, hashcat_logs = ?, cracked_handshake = ? WHERE uuid_user = ? AND uuid = ?",
		entities.HandshakeTableName,
	)
	if _, err = repo.dbUser.Exec(updateQuery,
		assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake, userUUID, handshakeUUID,
	); err != nil {
		return nil, err
	}

	// Fetch updated handshake
	results, err = qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName),
		handshakeBuilder,
		userUUID, handshakeUUID,
	)
	if err != nil {
		return nil, err
	}

	return results[0].(*entities.Handshake), nil
}

// UpdateClientTask updates client task (GRPC version)
func (repo *Repository) UpdateClientTask(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return repo.updateClientTaskCommon(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake, false)
}

// UpdateClientTaskRest updates client task (REST version)
func (repo *Repository) UpdateClientTaskRest(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake string) (*entities.Handshake, error) {
	return repo.updateClientTaskCommon(userUUID, handshakeUUID, assignedClientUUID, status, hashcatOptions, hashcatLogs, crackedHandshake, true)
}

// DeleteClient deletes a client record
func (repo *Repository) DeleteClient(userUUID, clientUUID string) (bool, error) {
	_, err := repo.dbUser.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE uuid_user = ? AND uuid = ?", entities.ClientTableName),
		userUUID, clientUUID,
	)
	return err == nil, err
}

// GetRaspberryPiByUserID returns paginated raspberry pi devices for a user
func (repo *Repository) GetRaspberryPiByUserID(userUUID string, offset uint) (rsps []*entities.RaspberryPI, length int, e error) {
	rspBuilder := func() (any, []any) {
		r := &entities.RaspberryPI{}
		return r, []any{
			&r.UserUUID,
			&r.RaspberryPIUUID,
			&r.MachineID,
			&r.EncryptionKey,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?",
			entities.RaspberryPiTableName, constants.Limit),
		rspBuilder,
		userUUID, (offset-1)*constants.Limit,
	)
	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		rsps = append(rsps, item.(*entities.RaspberryPI))
	}

	count, err := qq.countQueryResults(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ?", entities.RaspberryPiTableName),
		userUUID,
	)
	return rsps, count, err
}

// GetHandshakesByUserID returns paginated handshakes for a user
func (repo *Repository) GetHandshakesByUserID(userUUID string, offset uint) (handshakes []*entities.Handshake, length int, e error) {
	handshakeBuilder := func() (any, []any) {
		h := &entities.Handshake{}
		return h, []any{
			&h.UserUUID,
			&h.ClientUUID,
			&h.UUID,
			&h.SSID,
			&h.BSSID,
			&h.UploadedDate,
			&h.Status,
			&h.CrackedDate,
			&h.HashcatOptions,
			&h.HashcatLogs,
			&h.CrackedHandshake,
			&h.HandshakePCAP,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? LIMIT %v OFFSET ?",
			entities.HandshakeTableName, constants.Limit),
		handshakeBuilder,
		userUUID, (offset-1)*constants.Limit,
	)
	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		handshakes = append(handshakes, item.(*entities.Handshake))
	}

	count, err := qq.countQueryResults(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid_user = ?", entities.HandshakeTableName),
		userUUID,
	)
	return handshakes, count, err
}

// GetHandshakesByStatus returns handshakes filtered by status
func (repo *Repository) GetHandshakesByStatus(filterStatus string) (handshakes []*entities.Handshake, length int, e error) {
	handshakeBuilder := func() (any, []any) {
		h := &entities.Handshake{}
		return h, []any{
			&h.UserUUID,
			&h.ClientUUID,
			&h.UUID,
			&h.SSID,
			&h.BSSID,
			&h.UploadedDate,
			&h.Status,
			&h.CrackedDate,
			&h.HashcatOptions,
			&h.HashcatLogs,
			&h.CrackedHandshake,
			&h.HandshakePCAP,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE status = ?", entities.HandshakeTableName),
		handshakeBuilder,
		filterStatus,
	)
	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		handshakes = append(handshakes, item.(*entities.Handshake))
	}
	return handshakes, len(results), nil
}

// GetHandshakesByBSSIDAndSSID checks for existing handshake records
func (repo *Repository) GetHandshakesByBSSIDAndSSID(userUUID, bssid, ssid string) (handshakes []*entities.Handshake, length int, e error) {
	handshakeBuilder := func() (any, []any) {
		h := &entities.Handshake{}
		return h, []any{
			&h.UserUUID,
			&h.ClientUUID,
			&h.UUID,
			&h.SSID,
			&h.BSSID,
			&h.UploadedDate,
			&h.Status,
			&h.CrackedDate,
			&h.HashcatOptions,
			&h.HashcatLogs,
			&h.CrackedHandshake,
			&h.HandshakePCAP,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND bssid = ? AND ssid = ?", entities.HandshakeTableName),
		handshakeBuilder,
		userUUID, bssid, ssid,
	)
	if err != nil {
		return nil, -1, err
	}

	for _, item := range results {
		handshakes = append(handshakes, item.(*entities.Handshake))
	}
	return handshakes, len(results), nil
}

// CreateCertForClient generates and stores client certificates
func (repo *Repository) CreateCertForClient(userUUID, clientUUID string, clientCert, clientKey []byte) (string, error) {
	if repo.certs.caCert == nil || repo.certs.caKey == nil {
		return "", customErrors.ErrCertsNotInitialized
	}

	certID := uuid.New().String()
	_, err := repo.dbCerts.Exec(
		fmt.Sprintf("INSERT INTO %s(uuid, uuid_user, client_uuid, ca_cert, client_cert, client_key) VALUES(?,?,?,?,?,?)", entities.CertTableName),
		certID, userUUID, clientUUID, repo.certs.caCert, clientCert, clientKey,
	)
	return certID, err
}

// UpdateCerts updates client certificates during server startup
func (repo *Repository) UpdateCerts(client *entities.Client, caCert, clientCert, clientKey []byte) error {
	_, err := repo.dbCerts.Exec(
		fmt.Sprintf("UPDATE %s SET ca_cert = ?, client_cert = ?, client_key = ? WHERE client_uuid = ?", entities.CertTableName),
		caCert, clientCert, clientKey, client.ClientUUID,
	)
	return err
}

// UpdateUserPassword REST/API update user password
func (repo *Repository) UpdateUserPassword(userUUID, password string) error {
	psw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = repo.dbUser.Exec(
		fmt.Sprintf("UPDATE %s SET password = ? WHERE uuid = ?", entities.UserTableName),
		string(psw), userUUID,
	)
	return err
}

// CreateHandshake creates a new handshake record
func (repo *Repository) CreateHandshake(userUUID, ssid, bssid, status, handshakePcap string) (string, error) {
	handshakeID := uuid.New().String()
	_, err := repo.dbUser.Exec(
		fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, ssid, bssid, status, handshake_pcap) VALUES(?,?,?,?,?,?)", entities.HandshakeTableName),
		userUUID, handshakeID, ssid, bssid, status, handshakePcap,
	)
	return handshakeID, err
}

// CreateRaspberryPI creates a new raspberry pi device entry
func (repo *Repository) CreateRaspberryPI(userUUID, machineID, encryptionKey string) (string, error) {
	rspID := uuid.New().String()
	_, err := repo.dbUser.Exec(
		fmt.Sprintf("INSERT INTO %s(uuid_user, uuid, machine_id, encryption_key) VALUES(?,?,?,?)", entities.RaspberryPiTableName),
		userUUID, rspID, machineID, encryptionKey,
	)
	return rspID, err
}

// DeleteRaspberryPI deletes a raspberry pi record
func (repo *Repository) DeleteRaspberryPI(userUUID, rspUUID string) (bool, error) {
	_, err := repo.dbUser.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE uuid_user = ? AND uuid = ?", entities.RaspberryPiTableName),
		userUUID, rspUUID,
	)
	return err == nil, err
}

// DeleteHandshake deletes a handshake record
func (repo *Repository) DeleteHandshake(userUUID, handshakeUUID string) (bool, error) {
	_, err := repo.dbUser.Exec(
		fmt.Sprintf("DELETE FROM %s WHERE uuid_user = ? AND uuid = ?", entities.HandshakeTableName),
		userUUID, handshakeUUID,
	)
	return err == nil, err
}

// GetClientInfo retrieves client information by machine ID
func (repo *Repository) GetClientInfo(userUUID, machineID string) (*entities.Client, error) {
	clientBuilder := func() (any, []any) {
		c := &entities.Client{}
		return c, []any{
			&c.UserUUID,
			&c.ClientUUID,
			&c.Name,
			&c.LatestIP,
			&c.CreationTime,
			&c.LatestConnectionTime,
			&c.MachineID,
			&c.EnabledEncryption,
		}
	}

	qq := queryHandler{repo.dbUser}
	results, err := qq.queryEntities(
		fmt.Sprintf("SELECT * FROM %s WHERE uuid_user = ? AND machine_id = ?", entities.ClientTableName),
		clientBuilder,
		userUUID, machineID,
	)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, customErrors.ErrNoClientFound
	}
	return results[0].(*entities.Client), nil
}
