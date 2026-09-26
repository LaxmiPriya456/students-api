// Package types holds the data structures shared by the other packages.
//
// JAVA EQUIVALENT: your model / DTO classes.
package types

// Student is one student, as JSON and as a database row.
type Student struct {
	// Two struct tags per field:
	//   `json:"..."`             -> the JSON key name (like @JsonProperty)
	//   `validate:"required"`    -> read by go-playground/validator,
	//                               like @NotNull / @NotBlank in Java
	Id    int64  `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age" validate:"required"`
}
