# Mongotest

##### Mongodb wrapper of github.com/ory/dockertest

Testing software with mongodb docker container.

### Install
```
go get -u github.com/pokerblow/mongotest
```

### Using mongotest

```go
func TestMain(m *testing.M) {
	cc := mongotest.StartMongoContainer("4.2")
	// to get uri of mongodb write `cc.GetMongoURI("databaseName")`

	code := m.Run()

	cc.KillMongoContainer()

	os.Exit(code)
}
```
