# go-service-toolkit
A useful toolkit for creating web services in Go.

The majority of the patterns here were taken from and inspired by the excellent
[Let's Go](https://lets-go.alexedwards.net/) and [Let's Go
Further](https://lets-go-further.alexedwards.net/) books written by Alex
Edwards.

# Getting Started
## Project Setup
To get started using this toolkit, first, create a new, empty Go project and then run the following commands from the root directory of the project:

```bash
# Initialise a new Go project with the given MODULE_PATH.
go mod init MODULE_PATH

# Fetch the latest version of this package.
go get github.com/m5lapp/go-service-toolkit@latest

# Create the required directories.
mkdir -p cmd/api/ internal/data/ migrations/

# Create the skeleton files required.
touch .envrc cmd/api/{main.go,routes.go} internal/data/models.go

# Add the .envrc file to the .gitignore.
cat >> .gitignore << 'EOF'
# Development configuration files.
.envrc
```

## Models
In the `internal/data/` directory, create structs to represent any data models you need along with JSON struct tags if required. For each one, also create a corresponding `Model` struct containing a pointer to a `sql.DB`.

```go
package data

type Author struct {
    Name      string    `json:"name"`
    Birthdate time.Time `json:"birth_date"`
}

type Book struct {
    Title  string `json:"title"`
    Author Author `json:"author"`
}

type AuthorModel struct {
    DB *sql.DB
}

type AuthorModel struct {
    DB *sql.DB
}

```

Then, in the `internal/data/models.go` file, create a `Models` struct that contains an instance of each of models that you just defined, as well as a `NewModels()` function that returns an instance of the `Models` struct with the provided `*sql.DB` in each Model.

```go
package data

type Models struct {
    Author AuthorModel
    Book   BookModel
}

func NewModels(db *sql.DB) Models {
    return Models{
        Author: AuthorModel{DB: db},
        Book:   BookModel{DB: db},
    }
}
```

## Configuration
In the `cmd/api/main.go` file, create a struct that **embeds** a `webapp.WebApp` and also contains an instance of the `data.Models` struct from above. You can also add any of the command line flag structs from the `config` package if they fit your use case. The `config.SqlDB` struct in particular will be useful if you want to use a SQL database. Additionally, you can add any other dependencies you might need such as a `mailer.Mailer`.

```go
package main
type app struct {
    webapp.WebApp
    models  data.Models
    dbCfg   config.SqlDB
    smtpCfg config.Smtp
    mailer  mailer.Mailer
}
```

Next, create an instance of each config struct you are using and then call the `Flags()` method along with the default values for each one, then call `flag.Parse()`.
```go
package main

var serverCfg config.Server
var dbCfg config.SqlDB
var smtpCfg config.Smtp

serverCfg.Flags(":8080")
dbCfg.Flags("postgres", 25, 25, "15m")
smtpCfg.Flags("", "")

flag.Parse()
```

Finally, create a new `slog.Logger` and, if required, use the `sqldb.OpenDB()` method with the `config.SqlDB` struct as its parameter to create a new `sql.DB` instance.

Once that's done, create a new instance of the `app` struct you defined earlier with all the dependencies you have created. Use the `webapp.New()` function to instantiate the embeded `webapp.WebApp`.

Finally, call the `Serve()` method on it, passing in the `http.Handler` returned by the `routes()` function that we will define shortly.
```go
// Create a new slog.Logger instance.
logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
logger := slog.New(logHandler)

// If required, create a new sql.DB instance.
db, err := sqldb.OpenDB(dbCfg)
if err != nil {
    logger.Error(err.Error(), nil)
    os.Exit(1)
}
defer db.Close()

logger.Info("Database connection pool established")

// Create an instance of the app struct you defined, passing in the required
// configurations and dependencies.
app := &app{
    WebApp:  webapp.New(serverCfg, logger),
    models:  data.NewModels(db),
    cdCfg:   dbCfg,
    smtpCfg: smtpCfg,
    mailer:  mailer.New(&smtpCfg, templateFS),
}

// Call the app.Serve() method to start the application running.
err = app.Serve(app.routes())
if err != nil {
    logger.Error(err.Error(), nil)
    os.Exit(1)
}
```

## Set up the Routes
In the `cmd/api/routes.go` file, create a `routes()` method with the `app` struct as the receiver, add any routes you want to create to it and then return the `app.Router` struct wrapped in any middlewares from the `webapp.WebApp` that you want to utilise.

```go
package main

func (app *app) routes() http.Handler {
    app.Router.HandlerFunc(http.MethodPost, "/v1/authors", app.createAuthorHandler)
    app.Router.HandlerFunc(http.MethodPost, "/v1/books", app.createBookHandler)

    return app.Metrics(app.RecoverPanic(app.Router))
}
```

The handlers themselves should then be created in the main package under `cmd/api/`.

# Endpoints
Out of the box, the following endpoints are provided:

| Endpoint  | Method  | Description                                            |
| --------- | ------  | ------------------------------------------------------ |
| `/debug`  | GET     | Low-level application metrics and information          |
| `/health` | GET     | Health and status information                          |
