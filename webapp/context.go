package webapp

// import (
// 	"context"
// 	"net/http"

// 	"m5l.app/go-service-toolkit/data"
// )

// type contextKey string

// const userContextKey = contextKey("user")

// func (app *WebApp) contextSetUser(r *http.Request, user *data.User) *http.Request {
// 	ctx := context.WithValue(r.Context(), userContextKey, user)
// 	return r.WithContext(ctx)
// }

// func (app *WebApp) contextGetUser(r *http.Request) *data.User {
// 	user, ok := r.Context().Value(userContextKey).(*data.User)
// 	if !ok {
// 		panic("missing user value in request context")
// 	}

// 	return user
// }
