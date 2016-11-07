package models

import (
	"database/sql"
	"fmt"
)

type Instance struct {
	Id                int
	Host              string `json:"host"`
	Port              string `json:"port"`
	Instance_name     string `json:"instance_name"`
	Instance_username string `json:"instance_username"`
	Instance_password string `json:"instance_password"`
	Uri               string `json:"uri"`
	Username          string `json:"username"`
}

type createResult struct {
	Uri      string `json:"uri"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateInstance(db *sql.DB, instanceInfo *Instance) (*createResult, error) {
	logger.Info("Begin create a data instance model.")

	sqlstr := fmt.Sprintf(`insert into DF_DATA_INSTANCE (
				HSOTNAME, PORT, INSTANCE_NAME, INSTANCE_USERNAME, INSTANCE_PASSWORD, URI, USERNAME
				) values (?, ?, ?, ?, ?, ?, ?)`,
	)

	_, err := db.Exec(sqlstr,
		instanceInfo.Host, instanceInfo.Port, instanceInfo.Instance_name, instanceInfo.Instance_username,
		instanceInfo.Instance_password, instanceInfo.Uri, instanceInfo.Username,
	)
	if err != nil {
		logger.Error("Exec err : %v", err)
		return nil, err
	}

	result := createResult{Uri: instanceInfo.Uri, Hostname: instanceInfo.Host, Port: instanceInfo.Port,
		Name: instanceInfo.Instance_name, Username: instanceInfo.Instance_username,
		Password: instanceInfo.Instance_password}

	logger.Info("End create a data instance model.")
	return &result, err
}
