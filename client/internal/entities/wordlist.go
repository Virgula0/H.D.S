package entities

const WordlistTableName = "wordlist"

// Wordlist represents the table structure with constraints
type Wordlist struct {
	UUID                 string `db:"UUID" primary:"true"` // Primary Key
	UserUUID             string `db:"UUID_USER"`
	ClientUUID           string `db:"CLIENT_UUID"`
	WordlistName         string `db:"WORDLIST_NAME"`
	WordlistHash         string `db:"WORDLIST_HASH"`
	WordlistLines        int    `db:"WORDLIST_LINES"`
	WordlistLocationPath string `db:"WORDLIST_LOCATION_PATH" default:"'wordlists'"`
}
