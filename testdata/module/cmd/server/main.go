package main

import (
	_ "example.com/mod/internal/app"
	_ "example.com/mod/internal/store/postgres"
	_ "github.com/lllypuk/other/postgres"
)

func main() {}
