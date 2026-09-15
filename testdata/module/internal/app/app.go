// Package app — короткий док.
package app

import (
	"example.com/mod/internal/store/postgres"
	"fmt"
)

// Long — док на четыре строки,
// вторая,
// третья,
// четвёртая.
func Long() {
	// внутри тела три строки,
	// вторая,
	// третья.
	fmt.Println(postgres.Name)
}

// Holed — блок с пустой строкой.
//
// Хвост.
func Holed() {}

//go:generate echo directive
//nolint:all // директивы не считаются
func Fine() {}
