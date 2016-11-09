# datafoundry_data_instance

## 数据库设计
```
CREATE TABLE IF NOT EXISTS DF_DATA_INSTANCE
(
    INSTANCE_ID          BIGINT NOT NULL AUTO_INCREMENT,
    HSOTNAME             VARCHAR(128) NOT NULL,
    PORT                 VARCHAR(128) NOT NULL,
    INSTANCE_DATA	       VARCHAR(128) NOT NULL,
    INSTANCE_USERNAME    VARCHAR(128) NOT NULL,
    INSTANCE_PASSWORD    VARCHAR(128) NOT NULL,
    URI                  VARCHAR(256) NOT NULL,
    USERNAME             VARCHAR(64)  NOT NULL,
    CREATE_TIME          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UPDATE_TIME          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (INSTANCE_ID)
) DEFAULT CHARSET=UTF8;

CREATE TABLE IF NOT EXISTS DF_DATA_INSTANCE_SERVICE
(
    SERVICE_ID           BIGINT NOT NULL AUTO_INCREMENT,
    SERVICE_TYPE         VARCHAR(128) NOT NULL,
    SERVICE_DATA         VARCHAR(128) NOT NULL,
    SERVICE_ADDR         VARCHAR(256) NOT NULL,
    SERVICE_PORT         VARCHAR(10)  NOT NULL,
    USERNAME             VARCHAR(128) NOT NULL,
    PASSWORD             VARCHAR(128) NOT NULL,
    SERVICE_CLASS	       VARCHAR(128) NOT NULL,
    SERVICE_PROVIDER     VARCHAR(128) NOT NULL,
    DESCRIPTION          VARCHAR(1024) NOT NULL,
    IMAGEURL             VARCHAR(256) NOT NULL,
    CREATE_TIME          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UPDATE_TIME          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (SERVICE_ID)
) DEFAULT CHARSET=UTF8;
```

## API设计

### GET /integration/v1/instance?class={class}&provider={provider}

查询服务列表

Query Parameters:
```
class: 服务的类别。可选。如果忽略，表示所有类别。
provider: 提供方。可选。如果忽略，表示所有提供方。
page: 第几页。可选。最小值为1。默认为1。
size: 每页最多返回多少条数据。可选。最小为1，最大为100。默认为30。
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.total
data.results
data.results[0].class
data.results[0].provider
data.results[0].service_data
data.results[0].description
data.results[0].image_url
...
```

### POST /integration/v1/instance/{id}

创建一个数据集成服务

Path Parameters:
```
id: 服务id
```

Return Result (json):
```
code: 返回码
msg: 返回信息
data.uri
data.hostname
data.port
data.name
data.username
data.password
```

