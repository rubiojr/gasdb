module github.com/rubiojr/gasdb/_server

replace github.com/rubiojr/gasdb => ../

go 1.27

require (
	github.com/a-h/templ v0.3.1070
	github.com/go-chi/chi/v5 v5.3.2
	github.com/go-chi/httplog/v2 v2.1.1
	github.com/go-chi/httprate v0.16.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rubiojr/gasdb v1.5.0
	github.com/stretchr/testify v1.12.1
	github.com/tkrajina/gpxgo v1.5.1
)

require (
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/ncruces/go-sqlite3 v0.27.1 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
