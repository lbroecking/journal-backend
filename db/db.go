package db

import (
	"database/sql"
	"journal-backend/logging"

	_ "github.com/lib/pq"

	"os"
)

var Db *sql.DB

func init() {
	var err error
	logging.Log.Info("Conntecting to database...")
	//connStr := "dbname = postgres password = password2 user = postgres sslmode = disable host=shared-postgres  port= 5432"
	connStr := "postgresql://postgres.bopabjxclatablmbnwia:xumwap-zujquv-8pUbfo@aws-0-eu-central-1.pooler.supabase.com:6543/postgres"
	Db, err = sql.Open("postgres", connStr)
	if err != nil {
		logging.Log.Error("Failed to open database. Error: ", err)
		os.Exit(1)
	}
	logging.Log.Debug("Opened database.")

	if err = Db.Ping(); err != nil {
		logging.Log.Error("Failed to connect database. Error:", err)
		os.Exit(1)
	}

	_, err = Db.Exec("CREATE TABLE IF NOT EXISTS prekeybundle(userid varchar not null, registrationid integer not null, identitykey varchar not null, signedprekey varchar not null, signedprekey_id  integer not null, sig_signedprekey varchar not null, deviceid integer, username varchar, picture integer)")
	if err != nil {
		logging.Log.Error("Failed to create the table 'prekeybundle'. Error: ", err)
		os.Exit(1)
	}
	logging.Log.Debug("Created table 'prekeybundle' if table didn't exist.")

	_, err = Db.Exec("CREATE TABLE IF NOT EXISTS onetimeprekeys(opk varchar NOT NULL, opk_id integer NOT NULL, userid varchar NOT NULL)")
	if err != nil {
		logging.Log.Error("Failed to create the table 'onetimeprekeys'. Error: ", err)
		os.Exit(1)
	}
	logging.Log.Debug("Created table 'onetimeprekeys' if table didn't exist.")

	_, err = Db.Exec("CREATE TABLE IF NOT EXISTS messages(senderid varchar not null,	recipientid varchar not null, senderdeviceid integer not null, content varchar not null)")
	if err != nil {
		logging.Log.Error("Failed to create the table 'messages'. Error: ", err)
		os.Exit(1)
	}
	logging.Log.Debug("Created table 'messages' if table didn't exist.")

	logging.Log.Info("Conntected to database.")
}
