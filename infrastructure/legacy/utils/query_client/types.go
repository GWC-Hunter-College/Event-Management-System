package query_client

type databaseCredentials struct {
	user     string
	password string
	host     string
	port     string
	address  string
}

type Query struct {
	Filepath string
	Args     []any
}
