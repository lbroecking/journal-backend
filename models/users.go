package models

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"journal-backend/db"
	"journal-backend/logging"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type User struct {
	ID     int    `json:"id"`
	UserId string `json:"user_Id"`
	Name   string `json:"username"`
}

type UserModify struct {
	UserId   string `json:"user_id"`
	UserName string `json:"username"`
	Picture  int    `json:"avatar_url"`
}

type PreKeys struct {
	//UserId string `json:"userId"`
	IdentityKeyPublic string `json:"identityKeyPublic"`
	RegistrationID    int    `json:"registrationId"`
	SignedPreKey      struct {
		KeyID     int    `json:"keyId"`
		PublicKey string `json:"publicKey"`
		Signature string `json:"signature"`
	} `json:"signedPreKey"`
}

type OneTimePreKey struct {
	KeyID     int    `json:"keyId"`
	PublicKey string `json:"publicKey"`
}

type BundleWithOPKS struct {
	User
	PreKeys
	OneTimePreKeys []OneTimePreKey `json:"preKeys"`
}

func CreateNewPreKeyBundle(createUser BundleWithOPKS) (BundleWithOPKS, error) {
	//uid erstellen
	uid := uuid.NewV4()
	logging.Log.Debug("Generated a uuid for new user.")

	uuidString := fmt.Sprintf("%s", uid)

	//UserID hashen
	idHash := sha512.New()
	idHash.Write([]byte(uuidString))
	uuidHashed := idHash.Sum(nil)

	//erste 8 Zeichen von userIdHashed als UserId in prekeystruct zuweisen
	createUser.UserId = hex.EncodeToString(uuidHashed)[:8]
	logging.Log.Debug("Generated a UserId for new user: ", createUser.UserId)

	//PreKeys+userId in DB schreiben
	_, err := db.Db.Exec("INSERT INTO prekeybundle (userid, registrationid, identitykey, signedprekey, signedprekey_id, sig_signedprekey, deviceid) VALUES ($1, $2, $3, $4, $5, $6, $7)", createUser.UserId, createUser.RegistrationID, createUser.IdentityKeyPublic, createUser.SignedPreKey.PublicKey, createUser.SignedPreKey.KeyID, createUser.SignedPreKey.Signature)
	if err != nil {
		logging.Log.Error("Failed to insert PreKeys + UserId + DeviceId in database. Error: ", err)
	}
	logging.Log.Debug("Stored received users PreKeys + UserId + DeviceId in database.")

	//OneTimePreKeys durchlaufen und speichern
	for i := 0; i <= len(createUser.OneTimePreKeys)-1; i++ {
		_, err = db.Db.Exec("INSERT INTO onetimeprekeys (opk, opk_id, userid) VALUES ($1, $2, $3)", createUser.OneTimePreKeys[i].PublicKey, createUser.OneTimePreKeys[i].KeyID, createUser.UserId)
		if err != nil {
			logging.Log.Error("Failed to insert OneTimePreKeys in database. Error: ", err)
		}
	}
	logging.Log.Debug("Stored received users OneTimePreKeys in database.")

	return createUser, nil
}

func TestUser(c *gin.Context) {
	logging.Log.Info("Received GET-Request")
	logging.Log.Info("Selecting UserId + username of all users stored in database...")
	rows, err := db.Db.Query("SELECT id, username FROM profiles")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var profiles []User
	for rows.Next() {
		var p User
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		profiles = append(profiles, p)
	}
	c.JSON(http.StatusOK, profiles)
	logging.Log.Info("Selecting UserId + Name of all users stored in database was successful")
	logging.Log.Info("Sended response to client.")
}
