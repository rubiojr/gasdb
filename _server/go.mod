module github.com/rubiojr/gasdb/_server

replace github.com/rubiojr/gasdb => ../

godebug tlsrsakex=1

go 1.24.2

require (
	github.com/a-h/templ v0.3.865
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/httplog/v2 v2.1.1
	github.com/go-chi/httprate v0.15.0
	github.com/muesli/gominatim v0.1.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rubiojr/gasdb v0.0.0-00010101000000-000000000000
	github.com/tkrajina/gpxgo v1.4.0
)

require (
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/ncruces/go-sqlite3 v0.25.1 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
