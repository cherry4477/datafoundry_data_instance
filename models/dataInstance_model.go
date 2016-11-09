package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type Instance struct {
	Id                int
	Host              string `json:"host"`
	Port              string `json:"port"`
	Instance_data     string `json:"instance_name"`
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

type retrieveResult struct {
	Service_id    int    `json:"service_id"`
	Class         string `json:"class"`
	Provider      string `json:"provider"`
	Instance_data string `json:"instance_data"`
	Description   string `json:"description"`
	ImageUrl      string `json:"image_url"`
}

func CreateInstance(db *sql.DB, instanceInfo *Instance) (*createResult, error) {
	logger.Info("Begin create a data instance model.")

	sqlstr := fmt.Sprintf(`insert into DF_DATA_INSTANCE (
				HSOTNAME, PORT, INSTANCE_DATA, INSTANCE_USERNAME, INSTANCE_PASSWORD, URI, USERNAME
				) values (?, ?, ?, ?, ?, ?, ?)`,
	)

	_, err := db.Exec(sqlstr,
		instanceInfo.Host, instanceInfo.Port, instanceInfo.Instance_data, instanceInfo.Instance_username,
		instanceInfo.Instance_password, instanceInfo.Uri, instanceInfo.Username,
	)
	if err != nil {
		logger.Error("Exec err : %v", err)
		return nil, err
	}

	result := createResult{Uri: instanceInfo.Uri, Hostname: instanceInfo.Host, Port: instanceInfo.Port,
		Name: instanceInfo.Instance_data, Username: instanceInfo.Instance_username,
		Password: instanceInfo.Instance_password}

	logger.Info("End create a data instance model.")
	return &result, err
}

func QueryServices(db *sql.DB, class, provider, orderBy string, sortOrder bool, offset int64, limit int) (int64, []*retrieveResult, error) {
	logger.Info("Begin get coupon list model.")

	sqlParams := make([]interface{}, 0, 4)

	// ...

	sqlWhere := ""
	class = strings.ToLower(class)
	if class != "" {
		if sqlWhere == "" {
			sqlWhere = "SERVICE_CLASS = ?"
		} else {
			sqlWhere = sqlWhere + " and SERVICE_CLASS = ?"
		}
		sqlParams = append(sqlParams, class)
	}

	provider = strings.ToLower(provider)
	if provider != "" {
		if sqlWhere == "" {
			sqlWhere = "SERVICE_PROVIDER = ?"
		} else {
			sqlWhere = sqlWhere + " and SERVICE_PROVIDER = ?"
		}
		sqlParams = append(sqlParams, provider)
	}

	// ...

	switch strings.ToLower(orderBy) {
	default:
		orderBy = "SERVICE_ID"
		sortOrder = false
	case "createtime":
		orderBy = "CREATE_TIME"
	}

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrderText[sortOrder])

	// ...

	logger.Debug("sqlWhere=%v", sqlWhere)
	return getCouponList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
}

const (
	SortOrder_Asc  = "asc"
	SortOrder_Desc = "desc"
)

// true: asc
// false: desc
var sortOrderText = map[bool]string{true: "asc", false: "desc"}

func ValidateSortOrder(sortOrder string, defaultOrder bool) bool {
	switch strings.ToLower(sortOrder) {
	case SortOrder_Asc:
		return true
	case SortOrder_Desc:
		return false
	}

	return defaultOrder
}

func ValidateOrderBy(orderBy string) string {
	switch orderBy {
	case "createtime":
		return "CREATE_TIME"
	}

	return ""
}

func getCouponList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*retrieveResult, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryCouponsCount(db, sqlWhere, sqlParams...)
	logger.Debug("count: %v", count)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*retrieveResult{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)

	logger.Debug("sqlWhere=%v", sqlWhere)
	subs, err := queryCoupons(db, sqlWhere,
		fmt.Sprintf(`order by %s`, sqlSort),
		limit, offset, sqlParams...)

	return count, subs, err
}
func queryCoupons(db *sql.DB, sqlWhere, orderBy string, limit int, offset int64, sqlParams ...interface{}) ([]*retrieveResult, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}

	logger.Debug("sqlWhere=%v", sqlWhere)
	sqlWhereAll := ""
	if sqlWhere != "" {
		sqlWhereAll = fmt.Sprintf("WHERE %s", sqlWhere)
	} else {
		sqlWhereAll = fmt.Sprintf(" %s", sqlWhere)
	}

	sql_str := fmt.Sprintf(`select
					SERVICE_ID, SERVICE_DATA, SERVICE_CLASS, SERVICE_PROVIDER, DESCRIPTION, IMAGEURL
					from DF_DATA_INSTANCE_SERVICE
					%s %s
					limit %d
					%s
					`,
		sqlWhereAll,
		orderBy,
		limit,
		offset_str)
	rows, err := db.Query(sql_str, sqlParams...)

	logger.Info(">>> %v", sql_str)

	if err != nil {
		logger.Error("Query err : %v", err)
		return nil, err
	}
	defer rows.Close()

	services := make([]*retrieveResult, 0, 100)
	for rows.Next() {
		service := &retrieveResult{}
		err := rows.Scan(
			&service.Service_id, &service.Instance_data, &service.Class, &service.Provider,
			&service.Description, &service.ImageUrl,
		)
		if err != nil {
			logger.Error("Scan err : %v", err)
			return nil, err
		}
		//validateApp(s) // already done in scanAppWithRows
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Err : ", err)
		return nil, err
	}

	logger.Info("End get service list model.")
	return services, nil
}

func queryCouponsCount(db *sql.DB, sqlWhere string, sqlParams ...interface{}) (int64, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	count := int64(0)
	sql_str := fmt.Sprintf(`select COUNT(*) from DF_DATA_INSTANCE_SERVICE %s`, sql_where_all)
	logger.Debug(">>>\n"+
		"	%s", sql_str)
	logger.Debug("sqlParams: %v", sqlParams)
	err := db.QueryRow(sql_str, sqlParams...).Scan(&count)
	if err != nil {
		logger.Error("Scan err : %v", err)
		return 0, err
	}

	return count, err
}

func validateOffsetAndLimit(count int64, offset *int64, limit *int) {
	if *limit < 1 {
		*limit = 1
	}
	if *offset >= count {
		*offset = count - int64(*limit)
	}
	if *offset < 0 {
		*offset = 0
	}
	if *offset+int64(*limit) > count {
		*limit = int(count - *offset)
	}
}

type ServiceInfo struct {
	Address      string
	Port         string
	Username     string
	Password     string
	Service_data string
}

func GetServiceInfo(db *sql.DB, serviceId string) (*ServiceInfo, error) {
	id, err := strconv.Atoi(serviceId)
	if err != nil {
		logger.Error("Atoi err: %v.", err)
		return nil, err
	}
	sql := "select SERVICE_ADDR, SERVICE_PORT, USERNAME, PASSWORD, SERVICE_DATA from DF_DATA_INSTANCE_SERVICE where SERVICE_ID = ?"

	info := ServiceInfo{}
	err = db.QueryRow(sql, id).Scan(&info.Address, &info.Port, &info.Username, &info.Password, &info.Service_data)
	if err != nil {
		logger.Error("Scan err: %v.", err)
		return nil, err
	}
	logger.Debug("service info: %v.", info)

	return &info, nil
}
