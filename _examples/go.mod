module example

go 1.24.2

toolchain go1.24.4

godebug tlsrsakex=1

replace github.com/rubiojr/gasdb => ../

require github.com/rubiojr/gasdb v0.0.0-00010101000000-000000000000

require (
	github.com/tkrajina/gpxgo v1.4.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
