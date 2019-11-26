package broker

func cfmysqlcreds(c map[string]interface{}) map[string]interface{} {
	my := make(map[string]interface{})

	my["name"] = c["DBName"]
	my["host"] = c["EndpointAddress"]
	my["hostname"] = c["EndpointAddress"]
	my["port"] = c["Port"]
	my["password"] = c["MasterPassword"]
	my["username"] = c["MasterUsername"]
	my["jdbcUrl"] = "jdbc:mysql://" + my["host"].(string) + ":"
	my["jdbcUrl"] = my["jdbcUrl"].(string) + my["port"].(string) + "/"
	my["jdbcUrl"] = my["jdbcUrl"].(string) + my["name"].(string) + "?user="
	my["jdbcUrl"] = my["jdbcUrl"].(string) + my["username"].(string) + "&password="
	my["jdbcUrl"] = my["jdbcUrl"].(string) + my["password"].(string) + "&useSSL=false"
	my["uri"] = "mysql://" + my["username"].(string) + ":" + my["password"].(string)
	my["uri"] = my["uri"].(string) + "@" + my["host"].(string) + ":" + my["port"].(string)
	my["uri"] = my["uri"].(string) + "/" + my["name"].(string) + "?reconnect=true"

	return my
}

func cfpostgrecreds(c map[string]interface{}) map[string]interface{} {
	p := make(map[string]interface{})

	p["db_host"] = c["EndpointAddress"]
	p["db_name"] = c["DBName"]
	p["db_port"] = c["Port"]
	p["host"] = c["EndpointAddress"]
	p["hostname"] = c["EndpointAddress"]
	p["jdbc_uri"] = "jdbc:postgresql://" + p["host"].(string) + ":"
	p["jdbc_uri"] = p["jdbc_uri"].(string) + p["db_port"].(string) + "/"
	p["jdbc_uri"] = p["jdbc_uri"].(string) + p["db_name"].(string)
	p["jdbc_read_uri"] = p["jdbc_uri"]
	p["password"] = c["MasterPassword"]
	p["port"] = c["Port"]
	p["read_host"] = c["EndpointAddress"]
	p["read_port"] = c["Port"]
	p["uri"] = "postgresql://" + c["MasterUsername"].(string) + ":"
	p["uri"] = p["uri"].(string) + c["MasterPassword"].(string) +"@"
	p["uri"] = p["uri"].(string) + c["EndpointAddress"].(string) + ":"
	p["uri"] = p["uri"].(string) + c["Port"].(string) + "/" + c["DBName"].(string)
	p["read_uri"] = p["uri"]
	p["username"] = c["MasterUsername"]

	return p
}

func cfs3creds(c map[string]interface{}) map[string]interface{} {
	c["access_key_id"] = c["S3AwsAccessKeyId"]
	c["secret_access_key"] = c["S3AwsSecretAccessKey"]
	c["arn"] = c["BucketArn"]

	return c
}