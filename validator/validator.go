package validator

import "regexp"

var (
	EmailRX    = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	UsernameRX = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9_-]{3,30}[A-Za-z0-9]")
)

// A Validator is simply a map from field names to error messages.
type Validator struct {
	Errors map[string]string
}

// New returns a pointer to a new, empty validator struct.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid determines if any errors have been picked up by the Validator.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds a single, new validation error to the Errors map. If an error
// with the same key already exists, then this is a no-op.
func (v *Validator) AddError(key, message string) {
	_, exists := v.Errors[key]

	if !exists {
		v.Errors[key] = message
	}
}

// Check will add a new error to the Errors map if ok evaluates to false. Any
// validation function or boolean operation can be evaluated for the ok
// parameter.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.Errors[key] = message
	}
}

// PermittedValue checks that the comparable value of type T is one of the
// provided permittedValues.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if permittedValues[i] == value {
			return true
		}
	}
	return false
}

// Matches checks if the given value satisfies the given regex rx.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique checks that the given slice of comparable values of type T contains
// only unique values.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// ValidateEmail performs standard validation checks against the provided email
// address and populates any errors into v.
func ValidateEmail(v *Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(Matches(email, EmailRX), "email", "must be a valid email address")
}
