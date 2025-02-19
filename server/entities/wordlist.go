package entities

const WordlistTableName = "wordlist"

type Wordlist struct {
	UUID                 string `db:"uuid"`
	UserUUID             string `db:"UUID_USER"`
	ClientUUID           string `db:"CLIENT_UUID"`
	WordlistName         string `db:"WORDLIST_NAME"`
	WordlistHash         string `db:"WORDLIST_HASH"`
	WordlistSize         int64  `db:"WORDLIST_SIZE"`
	WordlistFileContent  []byte `db:"FILE_CONTENT"`
	WordlistLocationPath string `db:"WORDLIST_LOCATION_PATH"`
}
